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

type EncodeS3 struct {
	Account                string                 `yaml:"account" json:"account" doc:"tenant id for this flow collector"`
	Endpoint               string                 `yaml:"endpoint" json:"endpoint" doc:"address of s3 server"`
	AccessKeyID            string                 `yaml:"accessKeyId" json:"accessKeyId" doc:"username to connect to server"`
	SecretAccessKey        string                 `yaml:"secretAccessKey" json:"secretAccessKey" doc:"password to connect to server"`
	Bucket                 string                 `yaml:"bucket" json:"bucket" doc:"bucket into which to store objects"`
	WriteTimeout           Duration               `yaml:"writeTimeout,omitempty" json:"writeTimeout,omitempty" doc:"timeout (in seconds) for write operation"`
	BatchSize              int                    `yaml:"batchSize,omitempty" json:"batchSize,omitempty" doc:"limit on how many flows will be buffered before being sent to an object"`
	Secure                 bool                   `yaml:"secure,omitempty" json:"secure,omitempty" doc:"true for https, false for http (default: false)"`
	ObjectHeaderParameters map[string]interface{} `yaml:"objectHeaderParameters,omitempty" json:"objectHeaderParameters,omitempty" doc:"parameters to include in object header (key/value pairs)"`
	// TBD: (TLS?) security parameters
	// TLS                    *ClientTLS             `yaml:"tls" json:"tls" doc:"TLS client configuration (optional)"`
}
