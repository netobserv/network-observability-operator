package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/netobserv/flowlogs-pipeline/pkg/operational"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/encode"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/encode/opentelemetry"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/extract"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/extract/conntrack"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/ingest"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/transform"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/utils"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/write"
	k8sutils "github.com/netobserv/flowlogs-pipeline/pkg/utils"
	"github.com/netobserv/gopipes/pkg/node"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	defaultNodeBufferLen          = 1000
	defaultExtractBatching        = 1000
	defaultExtractBatchingTimeout = 5 * time.Second
)

// Error wraps any error caused by a wrong formation of the pipeline
type Error struct {
	StageName string
	wrapped   error
}

func (e *Error) Error() string {
	return fmt.Sprintf("pipeline stage %q: %s", e.StageName, e.wrapped.Error())
}

func (e *Error) Unwrap() error {
	return e.wrapped
}

// builder stores the information that is only required during the build of the pipeline
type builder struct {
	pipelineStages   []*pipelineEntry
	configStages     []config.Stage
	configParams     []config.StageParam
	pipelineEntryMap map[string]*pipelineEntry
	createdStages    map[string]interface{}
	startNodes       []*node.Start[config.GenericMap]
	terminalNodes    []*node.Terminal[config.GenericMap]
	opMetrics        *operational.Metrics
	stageDuration    *prometheus.HistogramVec
	batchMaxLen      int
	batchTimeout     time.Duration
	nodeBufferLen    int
	updtChans        map[string]chan config.StageParam
}

type pipelineEntry struct {
	stageName   string
	stageType   string
	Ingester    ingest.Ingester
	Transformer transform.Transformer
	Extractor   extract.Extractor
	Encoder     encode.Encoder
	Writer      write.Writer
}

func getDynConfig(cfg *config.ConfigFileStruct) ([]config.StageParam, error) {
	k8sconfig, err := k8sutils.LoadK8sConfig(cfg.DynamicParameters.KubeConfigPath)
	if err != nil {
		log.Errorf("Cannot get k8s config: %v", err)
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(k8sconfig)
	if err != nil {
		log.Errorf("Cannot init k8s config: %v", err)
		return nil, err
	}
	cm, err := clientset.CoreV1().ConfigMaps(cfg.DynamicParameters.Namespace).Get(context.TODO(), cfg.DynamicParameters.Name, metav1.GetOptions{})
	if err != nil {
		log.Errorf("Cannot get dynamic config: %v", err)
		return nil, err
	}
	rawConfig, ok := cm.Data[cfg.DynamicParameters.FileName]
	if !ok {
		log.Errorf("Cannot get file in configMap: %v", err)
		return nil, err
	}
	dynConfig := config.HotReloadStruct{}
	err = json.Unmarshal([]byte(rawConfig), &dynConfig)
	if err != nil {
		log.Errorf("Cannot parse config: %v", err)
		return nil, err
	}
	return dynConfig.Parameters, nil
}

func newBuilder(cfg *config.ConfigFileStruct) *builder {
	// Get global metrics settings
	opMetrics := operational.NewMetrics(&cfg.MetricsSettings)
	stageDuration := opMetrics.GetOrCreateStageDurationHisto()

	bl := cfg.PerfSettings.BatcherMaxLen
	if bl == 0 {
		bl = defaultExtractBatching
	}
	bt := cfg.PerfSettings.BatcherTimeout
	if bt == 0 {
		bt = defaultExtractBatchingTimeout
	}
	nb := cfg.PerfSettings.NodeBufferLen
	if nb == 0 {
		nb = defaultNodeBufferLen
	}

	if cfg.DynamicParameters.Name != "" &&
		cfg.DynamicParameters.Namespace != "" &&
		cfg.DynamicParameters.FileName != "" {
		dynParameters, err := getDynConfig(cfg)
		if err == nil {
			cfg.Parameters = append(cfg.Parameters, dynParameters...)
		}
	}

	return &builder{
		pipelineEntryMap: map[string]*pipelineEntry{},
		createdStages:    map[string]interface{}{},
		configStages:     cfg.Pipeline,
		configParams:     cfg.Parameters,
		opMetrics:        opMetrics,
		stageDuration:    stageDuration,
		batchMaxLen:      bl,
		batchTimeout:     bt,
		nodeBufferLen:    nb,
		updtChans:        map[string]chan config.StageParam{},
	}
}

// use a preset ingester
func (b *builder) presetIngester(ing ingest.Ingester) {
	name := config.PresetIngesterStage
	log.Debugf("stage = %v", name)
	b.appendEntry(&pipelineEntry{
		stageName: name,
		stageType: StageIngest,
		Ingester:  ing,
	})
}

// read the configuration stages definition and instantiate the corresponding native Go objects
func (b *builder) readStages() error {
	for _, param := range b.configParams {
		log.Debugf("stage = %v", param.Name)
		pEntry := pipelineEntry{
			stageName: param.Name,
			stageType: findStageType(&param),
		}
		var err error
		switch pEntry.stageType {
		case StageIngest:
			pEntry.Ingester, err = getIngester(b.opMetrics, param)
		case StageTransform:
			pEntry.Transformer, err = getTransformer(b.opMetrics, param)
		case StageExtract:
			pEntry.Extractor, err = getExtractor(b.opMetrics, param)
		case StageEncode:
			pEntry.Encoder, err = getEncoder(b.opMetrics, param)
		case StageWrite:
			pEntry.Writer, err = getWriter(b.opMetrics, param)
		default:
			err = fmt.Errorf("invalid stage type: %v, stage name: %v", pEntry.stageType, pEntry.stageName)
		}
		if err != nil {
			return err
		}
		b.appendEntry(&pEntry)
	}
	log.Debugf("pipeline = %v", b.pipelineStages)
	return nil
}

func (b *builder) appendEntry(pEntry *pipelineEntry) {
	b.pipelineEntryMap[pEntry.stageName] = pEntry
	b.pipelineStages = append(b.pipelineStages, pEntry)
	log.Debugf("pipeline = %v", b.pipelineStages)
}

// reads the configured Go stages and connects between them
// readStages must be invoked before this
func (b *builder) build() (*Pipeline, error) {
	// accounts start and middle nodes that are connected to another node
	sendingNodes := map[string]struct{}{}
	// accounts middle or terminal nodes that receive data from another node
	receivingNodes := map[string]struct{}{}
	for _, connection := range b.configStages {
		if connection.Name == "" || connection.Follows == "" {
			// ignore entries that do not represent a connection
			continue
		}
		// instantiates (or loads from cache) the destination node of a connection
		dstEntry, ok := b.pipelineEntryMap[connection.Name]
		if !ok {
			return nil, fmt.Errorf("unknown pipeline stage: %s", connection.Name)
		}
		dstNode, err := b.getStageNode(dstEntry, connection.Name)
		if err != nil {
			return nil, err
		}
		dst, ok := dstNode.(node.Receiver[config.GenericMap])
		if !ok {
			return nil, fmt.Errorf("stage %q of type %q can't receive data",
				connection.Name, dstEntry.stageType)
		}
		// instantiates (or loads from cache) the source node of a connection
		srcEntry, ok := b.pipelineEntryMap[connection.Follows]
		if !ok {
			return nil, fmt.Errorf("unknown pipeline stage: %s", connection.Follows)
		}
		srcNode, err := b.getStageNode(srcEntry, connection.Follows)
		if err != nil {
			return nil, err
		}
		src, ok := srcNode.(node.Sender[config.GenericMap])
		if !ok {
			return nil, fmt.Errorf("stage %q of type %q can't send data",
				connection.Follows, srcEntry.stageType)
		}
		log.Infof("connecting stages: %s --> %s", connection.Follows, connection.Name)

		sendingNodes[connection.Follows] = struct{}{}
		receivingNodes[connection.Name] = struct{}{}
		// connects source and destination node, and catches any panic from the Go-Pipes library.
		var catchErr *Error
		func() {
			defer func() {
				if msg := recover(); msg != nil {
					catchErr = &Error{
						StageName: connection.Name,
						wrapped: fmt.Errorf("%q and %q stages haven't compatible input/outputs: %v",
							connection.Follows, connection.Name, msg),
					}
				}
			}()
			src.SendsTo(dst)
		}()
		if catchErr != nil {
			return nil, catchErr
		}
	}

	if err := b.verifyConnections(sendingNodes, receivingNodes); err != nil {
		return nil, err
	}
	if len(b.startNodes) == 0 {
		return nil, errors.New("no ingesters have been defined")
	}
	if len(b.terminalNodes) == 0 {
		return nil, errors.New("no writers have been defined")
	}
	return &Pipeline{
		startNodes:       b.startNodes,
		terminalNodes:    b.terminalNodes,
		pipelineStages:   b.pipelineStages,
		pipelineEntryMap: b.pipelineEntryMap,
		Metrics:          b.opMetrics,
	}, nil
}

// verifies that all the start and middle nodes send data to another node
// verifies that all the middle and terminal nodes receive data from another node
func (b *builder) verifyConnections(sendingNodes, receivingNodes map[string]struct{}) error {
	for _, stg := range b.pipelineStages {
		if isReceptor(stg) {
			if _, ok := receivingNodes[stg.stageName]; !ok {
				return &Error{
					StageName: stg.stageName,
					wrapped: fmt.Errorf("pipeline stage from type %q"+
						" should receive data from at least another stage", stg.stageType),
				}
			}
		}
		if isSender(stg) {
			if _, ok := sendingNodes[stg.stageName]; !ok {
				return &Error{
					StageName: stg.stageName,
					wrapped: fmt.Errorf("pipeline stage from type %q"+
						" should send data to at least another stage", stg.stageType),
				}
			}
		}
	}
	return nil
}

func isReceptor(p *pipelineEntry) bool {
	return p.stageType != StageIngest
}

func isSender(p *pipelineEntry) bool {
	return p.stageType != StageWrite && p.stageType != StageEncode
}

func (b *builder) runMeasured(name string, f func()) {
	start := time.Now()
	f()
	duration := time.Since(start)
	b.stageDuration.WithLabelValues(name).Observe(float64(duration.Milliseconds()))
}

func (b *builder) getStageNode(pe *pipelineEntry, stageID string) (interface{}, error) {
	if stg, ok := b.createdStages[stageID]; ok {
		return stg, nil
	}
	var stage interface{}
	// TODO: modify all the types' interfaces to not need to write loops here, the same
	// as we do with Ingest
	switch pe.stageType {
	case StageIngest:
		init := node.AsStart(pe.Ingester.Ingest)
		b.startNodes = append(b.startNodes, init)
		stage = init
	case StageWrite:
		term := node.AsTerminal(func(in <-chan config.GenericMap) {
			b.opMetrics.CreateInQueueSizeGauge(stageID, func() int { return len(in) })
			for i := range in {
				b.runMeasured(stageID, func() {
					pe.Writer.Write(i)
				})
			}
		}, node.ChannelBufferLen(b.nodeBufferLen))
		b.terminalNodes = append(b.terminalNodes, term)
		stage = term
	case StageEncode:
		encode := node.AsTerminal(func(in <-chan config.GenericMap) {
			b.opMetrics.CreateInQueueSizeGauge(stageID, func() int { return len(in) })
			for i := range in {
				b.runMeasured(stageID, func() {
					pe.Encoder.Encode(i)
				})
			}
		}, node.ChannelBufferLen(b.nodeBufferLen))
		b.terminalNodes = append(b.terminalNodes, encode)
		stage = encode
	case StageTransform:
		stage = node.AsMiddle(func(in <-chan config.GenericMap, out chan<- config.GenericMap) {
			b.opMetrics.CreateInQueueSizeGauge(stageID, func() int { return len(in) })
			b.opMetrics.CreateOutQueueSizeGauge(stageID, func() int { return len(out) })
			for i := range in {
				b.runMeasured(stageID, func() {
					if transformed, ok := pe.Transformer.Transform(i); ok {
						out <- transformed
					}
				})
			}
		}, node.ChannelBufferLen(b.nodeBufferLen))
	case StageExtract:
		stage = node.AsMiddle(func(in <-chan config.GenericMap, out chan<- config.GenericMap) {
			b.opMetrics.CreateInQueueSizeGauge(stageID, func() int { return len(in) })
			b.opMetrics.CreateOutQueueSizeGauge(stageID, func() int { return len(out) })
			// TODO: replace batcher by rewriting the different extractor implementations
			// to keep the status while processing flows one by one
			utils.Batcher(utils.ExitChannel(), b.batchMaxLen, b.batchTimeout, in,
				func(maps []config.GenericMap) {
					outs := pe.Extractor.Extract(maps)
					for _, o := range outs {
						out <- o
					}
				},
			)
		}, node.ChannelBufferLen(b.nodeBufferLen))
	default:
		return nil, &Error{
			StageName: stageID,
			wrapped:   fmt.Errorf("invalid stage type: %s", pe.stageType),
		}
	}
	b.createdStages[stageID] = stage
	return stage, nil
}

func getIngester(opMetrics *operational.Metrics, params config.StageParam) (ingest.Ingester, error) {
	var ingester ingest.Ingester
	var err error
	switch params.Ingest.Type {
	case api.FileType, api.FileLoopType, api.FileChunksType:
		ingester, err = ingest.NewIngestFile(params)
	case api.SyntheticType:
		ingester, err = ingest.NewIngestSynthetic(opMetrics, params)
	case api.CollectorType:
		ingester, err = ingest.NewIngestCollector(opMetrics, params)
	case api.StdinType:
		ingester, err = ingest.NewIngestStdin(opMetrics, params)
	case api.KafkaType:
		ingester, err = ingest.NewIngestKafka(opMetrics, params)
	case api.GRPCType:
		ingester, err = ingest.NewGRPCProtobuf(opMetrics, params)
	case api.FakeType:
		ingester, err = ingest.NewIngestFake(params)
	default:
		panic(fmt.Sprintf("`ingest` type %s not defined", params.Ingest.Type))
	}
	return ingester, err
}

func getWriter(opMetrics *operational.Metrics, params config.StageParam) (write.Writer, error) {
	var writer write.Writer
	var err error
	switch params.Write.Type {
	case api.GRPCType:
		writer, err = write.NewWriteGRPC(params)
	case api.StdoutType:
		writer, err = write.NewWriteStdout(params)
	case api.NoneType:
		writer, err = write.NewWriteNone()
	case api.LokiType:
		writer, err = write.NewWriteLoki(opMetrics, params)
	case api.IpfixType:
		writer, err = write.NewWriteIpfix(params)
	case api.FakeType:
		writer, err = write.NewWriteFake(params)
	default:
		panic(fmt.Sprintf("`write` type %s not defined; if no writer needed, specify `none`", params.Write.Type))
	}
	return writer, err
}

func getTransformer(opMetrics *operational.Metrics, params config.StageParam) (transform.Transformer, error) {
	var transformer transform.Transformer
	var err error
	switch params.Transform.Type {
	case api.GenericType:
		transformer, err = transform.NewTransformGeneric(params)
	case api.FilterType:
		transformer, err = transform.NewTransformFilter(params)
	case api.NetworkType:
		transformer, err = transform.NewTransformNetwork(params, opMetrics)
	case api.NoneType:
		transformer, err = transform.NewTransformNone()
	default:
		panic(fmt.Sprintf("`transform` type %s not defined; if no transformer needed, specify `none`", params.Transform.Type))
	}
	return transformer, err
}

func getExtractor(opMetrics *operational.Metrics, params config.StageParam) (extract.Extractor, error) {
	var extractor extract.Extractor
	var err error
	switch params.Extract.Type {
	case api.NoneType:
		extractor, _ = extract.NewExtractNone()
	case api.AggregateType:
		extractor, err = extract.NewExtractAggregate(params)
	case api.ConnTrackType:
		extractor, err = conntrack.NewConnectionTrack(opMetrics, params, clock.New())
	case api.TimebasedType:
		extractor, err = extract.NewExtractTimebased(params)
	default:
		panic(fmt.Sprintf("`extract` type %s not defined; if no extractor needed, specify `none`", params.Extract.Type))
	}
	return extractor, err
}

func getEncoder(opMetrics *operational.Metrics, params config.StageParam) (encode.Encoder, error) {
	var encoder encode.Encoder
	var err error
	switch params.Encode.Type {
	case api.PromType:
		encoder, err = encode.NewEncodeProm(opMetrics, params)
	case api.KafkaType:
		encoder, err = encode.NewEncodeKafka(opMetrics, params)
	case api.S3Type:
		encoder, err = encode.NewEncodeS3(opMetrics, params)
	case api.OtlpLogsType:
		encoder, err = opentelemetry.NewEncodeOtlpLogs(opMetrics, params)
	case api.OtlpMetricsType:
		encoder, err = opentelemetry.NewEncodeOtlpMetrics(opMetrics, params)
	case api.OtlpTracesType:
		encoder, err = opentelemetry.NewEncodeOtlpTraces(opMetrics, params)
	case api.NoneType:
		encoder, _ = encode.NewEncodeNone()
	default:
		panic(fmt.Sprintf("`encode` type %s not defined; if no encoder needed, specify `none`", params.Encode.Type))
	}
	return encoder, err
}

// findStageParameters finds the matching config.param structure and identifies the stage type
func findStageType(param *config.StageParam) string {
	log.Debugf("findStageType: stage = %v", param.Name)
	if param.Ingest != nil && param.Ingest.Type != "" {
		return StageIngest
	}
	if param.Transform != nil && param.Transform.Type != "" {
		return StageTransform
	}
	if param.Extract != nil && param.Extract.Type != "" {
		return StageExtract
	}
	if param.Encode != nil && param.Encode.Type != "" {
		return StageEncode
	}
	if param.Write != nil && param.Write.Type != "" {
		return StageWrite
	}
	return "unknown"
}
