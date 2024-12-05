/*
Copyright Agoda Services Co.,Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package logs

import (
	"context"
	"go.opentelemetry.io/otel"
	"log"
	"sync"
)

type simpleLogRecordProcessor struct {
	exporterMu sync.Mutex
	stopOnce   sync.Once
	exporter   LogRecordExporter
}

func (lrp *simpleLogRecordProcessor) Shutdown(ctx context.Context) error {
	return nil
}

func (lrp *simpleLogRecordProcessor) ForceFlush(ctx context.Context) error {
	return nil
}

var _ LogRecordProcessor = (*simpleLogRecordProcessor)(nil)

// NewSimpleLogRecordProcessor returns a new LogRecordProcessor that will synchronously
// send completed logs to the exporter immediately.
//
// This LogRecordProcessor is not recommended for production use. The synchronous
// nature of this LogRecordProcessor make it good for testing, debugging, or
// showing examples of other feature, but it will be slow and have a high
// computation resource usage overhead. The BatchLogsProcessor is recommended
// for production use instead.
func NewSimpleLogRecordProcessor(exporter LogRecordExporter) LogRecordProcessor {
	slp := &simpleLogRecordProcessor{
		exporter: exporter,
	}
	log.Printf("SimpleLogsProcessor is not recommended for production use, consider using BatchLogRecordProcessor instead.")

	return slp
}

// OnEmit Process immediately emits a LogRecord
func (lrp *simpleLogRecordProcessor) OnEmit(rol ReadableLogRecord) {
	lrp.exporterMu.Lock()
	defer lrp.exporterMu.Unlock()

	if err := lrp.exporter.Export(context.Background(), []ReadableLogRecord{rol}); err != nil {
		otel.Handle(err)
	}
}

// MarshalLog is the marshaling function used by the logging system to represent this LogRecord Processor.
func (lrp *simpleLogRecordProcessor) MarshalLog() interface{} {
	return struct {
		Type     string
		Exporter LogRecordExporter
	}{
		Type:     "SimpleLogRecordProcessor",
		Exporter: lrp.exporter,
	}
}
