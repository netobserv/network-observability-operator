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

package otel

import (
	"github.com/agoda-com/opentelemetry-logs-go/internal/global"
	"github.com/agoda-com/opentelemetry-logs-go/logs"
)

// GetLoggerProvider returns the registered global logger provider.
// If none is registered then an instance of NoopLoggerProvider is returned.
//
// loggerProvider := otel.GetLoggerProvider()
func GetLoggerProvider() logs.LoggerProvider {
	return global.LoggerProvider()
}

// SetLoggerProvider registers `lp` as the global logger provider.
func SetLoggerProvider(lp logs.LoggerProvider) {
	global.SetLoggerProvider(lp)
}
