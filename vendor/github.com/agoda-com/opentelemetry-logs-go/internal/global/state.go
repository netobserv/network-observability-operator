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
	"errors"
	"github.com/agoda-com/opentelemetry-logs-go/logs"
	"sync"
	"sync/atomic"
)

type (
	loggerProviderHolder struct {
		lp logs.LoggerProvider
	}
)

var (
	globalOtelLogger = defaultLoggerValue()

	delegateLoggerOnce sync.Once
)

// LoggerProvider is the internal implementation for global.LoggerProvider.
func LoggerProvider() logs.LoggerProvider {
	return globalOtelLogger.Load().(loggerProviderHolder).lp
}

// SetLoggerProvider is the internal implementation for global.SetLoggerProvider.
func SetLoggerProvider(lp logs.LoggerProvider) {
	current := LoggerProvider()

	if _, cOk := current.(*loggerProvider); cOk {
		if _, tpOk := lp.(*loggerProvider); tpOk && current == lp {
			// Do not assign the default delegating LoggerProvider to delegate
			// to itself.
			Error(
				errors.New("no delegate configured in logger provider"),
				"Setting logger provider to it's current value. No delegate will be configured",
			)
			return
		}
	}

	globalOtelLogger.Store(loggerProviderHolder{lp: lp})
}

func defaultLoggerValue() *atomic.Value {
	v := &atomic.Value{}
	v.Store(loggerProviderHolder{lp: &loggerProvider{}})
	return v
}
