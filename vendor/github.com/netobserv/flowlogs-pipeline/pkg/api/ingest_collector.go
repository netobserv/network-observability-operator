/*
 * Copyright (C) 2022 IBM, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package api

type IngestCollector struct {
	HostName    string `yaml:"hostName,omitempty" json:"hostName,omitempty" doc:"the hostname to listen on"`
	Port        int    `yaml:"port,omitempty" json:"port,omitempty" doc:"the port number to listen on, for IPFIX/NetFlow v9. Omit or set to 0 to disable IPFIX/NetFlow v9 ingestion"`
	PortLegacy  int    `yaml:"portLegacy,omitempty" json:"portLegacy,omitempty" doc:"the port number to listen on, for legacy NetFlow v5. Omit or set to 0 to disable NetFlow v5 ingestion"`
	BatchMaxLen int    `yaml:"batchMaxLen,omitempty" json:"batchMaxLen,omitempty" doc:"the number of accumulated flows before being forwarded for processing"`
}
