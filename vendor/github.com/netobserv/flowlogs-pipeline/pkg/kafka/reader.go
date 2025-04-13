package kafka

import (
	"errors"
	"os"
	"time"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/utils"
	kafkago "github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

var klog = logrus.WithField("component", "kafka-reader")

const defaultBatchReadTimeout = int64(1000)
const defaultKafkaBatchMaxLength = 500
const defaultKafkaCommitInterval = 500

func NewReader(config *api.IngestKafka) (*kafkago.Reader, int, error) {
	startOffsetString := config.StartOffset
	var startOffset int64
	switch startOffsetString {
	case "FirstOffset", "":
		startOffset = kafkago.FirstOffset
	case "LastOffset":
		startOffset = kafkago.LastOffset
	default:
		startOffset = kafkago.FirstOffset
		klog.Errorf("illegal value for StartOffset: %s; using default\n", startOffsetString)
	}
	klog.Debugf("startOffset = %v", startOffset)
	groupBalancers := make([]kafkago.GroupBalancer, 0)
	for _, gb := range config.GroupBalancers {
		switch gb {
		case "range":
			groupBalancers = append(groupBalancers, &kafkago.RangeGroupBalancer{})
		case "roundRobin":
			groupBalancers = append(groupBalancers, &kafkago.RoundRobinGroupBalancer{})
		case "rackAffinity":
			groupBalancers = append(groupBalancers, &kafkago.RackAffinityGroupBalancer{})
		default:
			klog.Warningf("groupbalancers parameter missing")
			groupBalancers = append(groupBalancers, &kafkago.RoundRobinGroupBalancer{})
		}
	}

	batchReadTimeout := defaultBatchReadTimeout
	if config.BatchReadTimeout != 0 {
		batchReadTimeout = config.BatchReadTimeout
	}
	klog.Debugf("batchReadTimeout = %d", batchReadTimeout)

	commitInterval := int64(defaultKafkaCommitInterval)
	if config.CommitInterval != 0 {
		commitInterval = config.CommitInterval
	}
	klog.Debugf("commitInterval = %d", config.CommitInterval)

	dialer := &kafkago.Dialer{
		Timeout:   kafkago.DefaultDialer.Timeout,
		DualStack: kafkago.DefaultDialer.DualStack,
	}
	if config.TLS != nil {
		klog.Infof("Using TLS configuration: %v", config.TLS)
		tlsConfig, err := config.TLS.Build()
		if err != nil {
			return nil, 0, err
		}
		dialer.TLS = tlsConfig
	}

	if config.SASL != nil {
		m, err := utils.SetupSASLMechanism(config.SASL)
		if err != nil {
			return nil, 0, err
		}
		dialer.SASLMechanism = m
	}

	readerConfig := kafkago.ReaderConfig{
		Brokers:        config.Brokers,
		Topic:          config.Topic,
		GroupID:        config.GroupID,
		GroupBalancers: groupBalancers,
		StartOffset:    startOffset,
		CommitInterval: time.Duration(commitInterval) * time.Millisecond,
		Dialer:         dialer,
	}

	if readerConfig.GroupID == "" {
		// Use hostname
		readerConfig.GroupID = os.Getenv("HOSTNAME")
	}

	if config.PullQueueCapacity > 0 {
		readerConfig.QueueCapacity = config.PullQueueCapacity
	}

	if config.PullMaxBytes > 0 {
		readerConfig.MaxBytes = config.PullMaxBytes
	}

	bml := defaultKafkaBatchMaxLength
	if config.BatchMaxLen != 0 {
		bml = config.BatchMaxLen
	}

	klog.Debugf("reader config: %#v", readerConfig)

	kafkaReader := kafkago.NewReader(readerConfig)
	if kafkaReader == nil {
		return nil, 0, errors.New("NewIngestKafka: failed to create kafka-go reader")
	}

	return kafkaReader, bml, nil
}
