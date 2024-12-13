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

package env

import (
	"os"
	"strconv"
)

// Environment variable names.
const (
	// BatchLogsProcessorScheduleDelayKey is the delay interval between two
	// consecutive exports (i.e. 5000).
	BatchLogsProcessorScheduleDelayKey = "OTEL_BLRP_SCHEDULE_DELAY"
	// BatchLogsProcessorExportTimeoutKey is the maximum allowed time to
	// export data (i.e. 3000).
	BatchLogsProcessorExportTimeoutKey = "OTEL_BLRP_EXPORT_TIMEOUT"
	// BatchLogsProcessorMaxQueueSizeKey is the maximum queue size (i.e. 2048).
	BatchLogsProcessorMaxQueueSizeKey = "OTEL_BLRP_MAX_QUEUE_SIZE"
	// BatchLogsProcessorMaxExportBatchSizeKey is the maximum batch size (i.e.
	// 512). Note: it must be less than or equal to
	// EnvBatchLogsProcessorMaxQueueSize.
	BatchLogsProcessorMaxExportBatchSizeKey = "OTEL_BLRP_MAX_EXPORT_BATCH_SIZE"
)

// firstInt returns the value of the first matching environment variable from
// keys. If the value is not an integer or no match is found, defaultValue is
// returned.
func firstInt(defaultValue int, keys ...string) int {
	for _, key := range keys {
		value := os.Getenv(key)
		if value == "" {
			continue
		}

		intValue, err := strconv.Atoi(value)
		if err != nil {
			//envconfig.Info("Got invalid value, number value expected.", key, value)
			return defaultValue
		}

		return intValue
	}

	return defaultValue
}

// IntEnvOr returns the int value of the environment variable with name key if
// it exists, it is not empty, and the value is an int. Otherwise, defaultValue is returned.
func IntEnvOr(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		//global.Info("Got invalid value, number value expected.", key, value)
		return defaultValue
	}

	return intValue
}

// BatchLogsProcessorScheduleDelay returns the environment variable value for
// the OTEL_BLRP_SCHEDULE_DELAY key if it exists, otherwise defaultValue is
// returned.
func BatchLogsProcessorScheduleDelay(defaultValue int) int {
	return IntEnvOr(BatchLogsProcessorScheduleDelayKey, defaultValue)
}

// BatchLogsProcessorExportTimeout returns the environment variable value for
// the OTEL_BLRP_EXPORT_TIMEOUT key if it exists, otherwise defaultValue is
// returned.
func BatchLogsProcessorExportTimeout(defaultValue int) int {
	return IntEnvOr(BatchLogsProcessorExportTimeoutKey, defaultValue)
}

// BatchLogsProcessorMaxQueueSize returns the environment variable value for
// the OTEL_BLRP_MAX_QUEUE_SIZE key if it exists, otherwise defaultValue is
// returned.
func BatchLogsProcessorMaxQueueSize(defaultValue int) int {
	return IntEnvOr(BatchLogsProcessorMaxQueueSizeKey, defaultValue)
}

// BatchLogsProcessorMaxExportBatchSize returns the environment variable value for
// the OTEL_BLRP_MAX_EXPORT_BATCH_SIZE key if it exists, otherwise defaultValue
// is returned.
func BatchLogsProcessorMaxExportBatchSize(defaultValue int) int {
	return IntEnvOr(BatchLogsProcessorMaxExportBatchSizeKey, defaultValue)
}
