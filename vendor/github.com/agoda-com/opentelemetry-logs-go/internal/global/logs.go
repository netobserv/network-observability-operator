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

package global

import (
	"github.com/agoda-com/opentelemetry-logs-go/logs"
	"sync"
	"sync/atomic"
)

// loggerProvider is a placeholder for a configured SDK LoggerProvider.
//
// All LoggerProvider functionality is forwarded to a delegate once
// configured.
type loggerProvider struct {
	mtx      sync.Mutex
	loggers  map[il]*logger
	delegate logs.LoggerProvider
}

// Compile-time guarantee that loggerProvider implements the LoggerProvider
// interface.
var _ logs.LoggerProvider = &loggerProvider{}

// setDelegate configures p to delegate all LoggerProvider functionality to
// provider.
//
// All Loggers provided prior to this function call are switched out to be
// Loggers provided by provider.
//
// It is guaranteed by the caller that this happens only once.
func (p *loggerProvider) setDelegate(provider logs.LoggerProvider) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	p.delegate = provider

	if len(p.loggers) == 0 {
		return
	}

	for _, t := range p.loggers {
		t.setDelegate(provider)
	}

	p.loggers = nil
}

// Logger implements LoggerProvider.
func (p *loggerProvider) Logger(name string, opts ...logs.LoggerOption) logs.Logger {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	if p.delegate != nil {
		return p.delegate.Logger(name, opts...)
	}

	// At this moment it is guaranteed that no sdk is installed, save the logger in the loggers map.

	c := logs.NewLoggerConfig(opts...)
	key := il{
		name:    name,
		version: c.InstrumentationVersion(),
	}

	if p.loggers == nil {
		p.loggers = make(map[il]*logger)
	}

	if val, ok := p.loggers[key]; ok {
		return val
	}

	t := &logger{name: name, opts: opts, provider: p}
	p.loggers[key] = t
	return t
}

type il struct {
	name    string
	version string
}

// logger is a placeholder for a logs.Logger.
//
// All Logger functionality is forwarded to a delegate once configured.
// Otherwise, all functionality is forwarded to a NoopLogger.
type logger struct {
	name     string
	opts     []logs.LoggerOption
	provider *loggerProvider

	delegate atomic.Value
}

// Compile-time guarantee that logger implements the logs.Logger interface.
var _ logs.Logger = &logger{}

func (t *logger) Emit(logRecord logs.LogRecord) {
	delegate := t.delegate.Load()
	if delegate != nil {
		delegate.(logs.Logger).Emit(logRecord)
	}
}

// setDelegate configures t to delegate all Logger functionality to Loggers
// created by provider.
//
// All subsequent calls to the Logger methods will be passed to the delegate.
//
// It is guaranteed by the caller that this happens only once.
func (t *logger) setDelegate(provider logs.LoggerProvider) {
	t.delegate.Store(provider.Logger(t.name, t.opts...))
}
