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

package logs // Package logs import "github.com/agoda-com/opentelemetry-logs-go/sdk/logs"

import (
	"context"
	"fmt"
	"github.com/agoda-com/opentelemetry-logs-go/internal/global"
	"github.com/agoda-com/opentelemetry-logs-go/logs"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/resource"
	"sync"
	"sync/atomic"
)

const (
	defaultLoggerName = "github.com/agoda-com/opentelemetry-logs-go/sdk/logs/provider"
)

// loggerProviderConfig Configuration for Logger Provider
type loggerProviderConfig struct {
	processors []LogRecordProcessor
	// resource contains attributes representing an entity that produces telemetry.
	resource *resource.Resource
}

// LoggerProviderOption configures a LoggerProvider.
type LoggerProviderOption interface {
	apply(loggerProviderConfig) loggerProviderConfig
}
type loggerProviderOptionFunc func(loggerProviderConfig) loggerProviderConfig

func (fn loggerProviderOptionFunc) apply(cfg loggerProviderConfig) loggerProviderConfig {
	return fn(cfg)
}

// WithLogRecordProcessor will configure processor to process logs
func WithLogRecordProcessor(logsProcessor LogRecordProcessor) LoggerProviderOption {
	return loggerProviderOptionFunc(func(cfg loggerProviderConfig) loggerProviderConfig {
		cfg.processors = append(cfg.processors, logsProcessor)
		return cfg
	})
}

// WithSyncer registers the exporter with the LoggerProvider using a
// SimpleLogRecordProcessor.
//
// This is not recommended for production use. The synchronous nature of the
// SimpleLogRecordProcessor that will wrap the exporter make it good for testing,
// debugging, or showing examples of other feature, but it will be slow and
// have a high computation resource usage overhead. The WithBatcher option is
// recommended for production use instead.
func WithSyncer(e LogRecordExporter) LoggerProviderOption {
	return WithLogRecordProcessor(NewSimpleLogRecordProcessor(e))
}

// WithBatcher registers the exporter with the LoggerProvider using a
// BatchLogRecordProcessor configured with the passed opts.
func WithBatcher(e LogRecordExporter, opts ...BatchLogRecordProcessorOption) LoggerProviderOption {
	return WithLogRecordProcessor(NewBatchLogRecordProcessor(e, opts...))
}

// WithResource will configure OTLP logger with common resource attributes.
//
// Parameters:
// r (*resource.Resource) list of resources will be added to every log as resource level tags
func WithResource(r *resource.Resource) LoggerProviderOption {
	return loggerProviderOptionFunc(func(cfg loggerProviderConfig) loggerProviderConfig {
		var err error
		cfg.resource, err = resource.Merge(resource.Environment(), r)
		if err != nil {
			otel.Handle(err)
		}
		return cfg
	})
}

// LoggerProvider provide access to Logger. The API is not intended to be called by application developers directly.
// see https://opentelemetry.io/docs/specs/otel/logs/bridge-api/#loggerprovider
type LoggerProvider struct {
	mu          sync.Mutex
	namedLogger map[instrumentation.Scope]*logger
	//cfg loggerProviderConfig

	logProcessors atomic.Pointer[logRecordProcessorStates]
	isShutdown    atomic.Bool

	// These fields are not protected by the lock mu. They are assumed to be
	// immutable after creation of the LoggerProvider.
	resource *resource.Resource
}

var _ logs.LoggerProvider = &LoggerProvider{}

func (lp *LoggerProvider) Logger(name string, opts ...logs.LoggerOption) logs.Logger {

	if lp.isShutdown.Load() {
		return logs.NewNoopLoggerProvider().Logger(name, opts...)
	}

	c := logs.NewLoggerConfig(opts...)

	if name == "" {
		name = defaultLoggerName
	}

	is := instrumentation.Scope{
		Name:      name,
		Version:   c.InstrumentationVersion(),
		SchemaURL: c.SchemaURL(),
	}

	t, ok := func() (logs.Logger, bool) {
		lp.mu.Lock()
		defer lp.mu.Unlock()
		// Must check the flag after acquiring the mutex to avoid returning a valid logger if Shutdown() ran
		// after the first check above but before we acquired the mutex.
		if lp.isShutdown.Load() {
			return logs.NewNoopLoggerProvider().Logger(name, opts...), true
		}

		t, ok := lp.namedLogger[is]
		if !ok {
			t = &logger{
				provider:             lp,
				instrumentationScope: is,
			}
		}
		return t, ok
	}()
	if !ok {
		// This code is outside the mutex to not hold the lock while calling third party logging code:
		// - That code may do slow things like I/O, which would prolong the duration the lock is held,
		//   slowing down all tracing consumers.
		// - Logging code may be instrumented with logging and deadlock because it could try
		//   acquiring the same non-reentrant mutex.
		global.Info("Logger created", "name", name, "version", is.Version, "schemaURL", is.SchemaURL)

	}
	return t
}

var _ logs.LoggerProvider = &LoggerProvider{}

func NewLoggerProvider(opts ...LoggerProviderOption) *LoggerProvider {
	o := loggerProviderConfig{}

	o = applyLoggerProviderEnvConfigs(o)

	for _, opt := range opts {
		o = opt.apply(o)
	}

	o = ensureValidLoggerProviderConfig(o)

	lp := &LoggerProvider{
		namedLogger: make(map[instrumentation.Scope]*logger),
		resource:    o.resource,
	}

	global.Info("LoggerProvider created", "config", o)

	lrpss := make(logRecordProcessorStates, 0, len(o.processors))
	for _, lrp := range o.processors {
		lrpss = append(lrpss, newLogsProcessorState(lrp))
	}
	lp.logProcessors.Store(&lrpss)

	return lp

}

func (p *LoggerProvider) getLogRecordProcessorStates() logRecordProcessorStates {
	return *(p.logProcessors.Load())
}

func (p LoggerProvider) Shutdown(ctx context.Context) error {
	// This check prevents deadlocks in case of recursive shutdown.
	if p.isShutdown.Load() {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	// This check prevents calls after a shutdown has already been done concurrently.
	if !p.isShutdown.CompareAndSwap(false, true) { // did toggle?
		return nil
	}

	var retErr error
	for _, lrps := range p.getLogRecordProcessorStates() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var err error
		lrps.state.Do(func() {
			err = lrps.lp.Shutdown(ctx)
		})
		if err != nil {
			if retErr == nil {
				retErr = err
			} else {
				// Poor man's list of errors
				retErr = fmt.Errorf("%v; %v", retErr, err)
			}
		}
	}
	p.logProcessors.Store(&logRecordProcessorStates{})
	return retErr

}

// ForceFlush immediately exports all logs that have not yet been exported for
// all the registered log processors.
func (p *LoggerProvider) ForceFlush(ctx context.Context) error {
	lrpss := p.getLogRecordProcessorStates()
	if len(lrpss) == 0 {
		return nil
	}

	for _, lrps := range lrpss {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := lrps.lp.ForceFlush(ctx); err != nil {
			return err
		}
	}
	return nil
}

func applyLoggerProviderEnvConfigs(cfg loggerProviderConfig) loggerProviderConfig {
	for _, opt := range loggerProviderOptionsFromEnv() {
		cfg = opt.apply(cfg)
	}

	return cfg
}

func loggerProviderOptionsFromEnv() []LoggerProviderOption {
	var opts []LoggerProviderOption

	return opts
}

// ensureValidLoggerProviderConfig ensures that given LoggerProviderConfig is valid.
func ensureValidLoggerProviderConfig(cfg loggerProviderConfig) loggerProviderConfig {

	if cfg.resource == nil {
		cfg.resource = resource.Default()
	}

	return cfg
}
