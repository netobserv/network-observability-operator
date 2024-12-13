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

// NewNoopLoggerProvider returns an implementation of LoggerProvider that
// performs no operations. The Logger created from the returned
// LoggerProvider also perform no operations.
func NewNoopLoggerProvider() LoggerProvider {
	return noopLoggerProvider{}
}

type noopLoggerProvider struct{}

var _ LoggerProvider = noopLoggerProvider{}

func (p noopLoggerProvider) Logger(string, ...LoggerOption) Logger {
	return noopLogger{}
}

type noopLogger struct{}

var _ Logger = noopLogger{}

func (n noopLogger) Emit(logRecord LogRecord) {}
