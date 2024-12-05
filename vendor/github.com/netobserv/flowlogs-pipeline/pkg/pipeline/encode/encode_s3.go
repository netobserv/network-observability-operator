/*
 * Copyright (C) 2022 IBM, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *	 http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package encode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/netobserv/flowlogs-pipeline/pkg/operational"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/utils"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

const (
	flpS3Version     = "v0.1"
	defaultBatchSize = 10
)

var (
	defaultTimeOut = api.Duration{Duration: 60 * time.Second}
)

type encodeS3 struct {
	s3Params          api.EncodeS3
	s3Writer          s3WriteEntries
	recordsWritten    prometheus.Counter
	pendingEntries    []config.GenericMap
	mutex             *sync.Mutex
	expiryTime        time.Time
	exitChan          <-chan struct{}
	streamID          string
	intervalStartTime time.Time
	sequenceNumber    int64
}

type s3WriteEntries interface {
	putObject(bucket string, objectName string, object map[string]interface{}) error
}

type encodeS3Writer struct {
	s3Client *minio.Client
	s3Params *api.EncodeS3
}

// The mutex must be held when calling writeObject
func (s *encodeS3) writeObject() error {
	nLogs := len(s.pendingEntries)
	if nLogs > s.s3Params.BatchSize {
		nLogs = s.s3Params.BatchSize
	}
	now := time.Now()
	object := s.GenerateStoreHeader(s.pendingEntries[0:nLogs], s.intervalStartTime, now)
	year := fmt.Sprintf("%04d", now.Year())
	month := fmt.Sprintf("%02d", now.Month())
	day := fmt.Sprintf("%02d", now.Day())
	hour := fmt.Sprintf("%02d", now.Hour())
	seq := fmt.Sprintf("%08d", s.sequenceNumber)
	objectName := s.s3Params.Account + "/year=" + year + "/month=" + month + "/day=" + day + "/hour=" + hour + "/stream-id=" + s.streamID + "/" + seq
	log.Debugf("S3 writeObject: objectName = %s", objectName)
	log.Debugf("S3 writeObject: object = %v", object)
	s.pendingEntries = s.pendingEntries[nLogs:]
	s.intervalStartTime = now
	s.expiryTime = now.Add(s.s3Params.WriteTimeout.Duration)
	s.sequenceNumber++
	// send object to object store
	err := s.s3Writer.putObject(s.s3Params.Bucket, objectName, object)
	if err != nil {
		log.Errorf("error in writing object: %v", err)
	}
	return err
}

func (s *encodeS3) GenerateStoreHeader(flows []config.GenericMap, startTime time.Time, endTime time.Time) map[string]interface{} {
	object := make(map[string]interface{})
	// copy user defined keys from config to object header
	for key, value := range s.s3Params.ObjectHeaderParameters {
		object[key] = value
	}
	object["version"] = flpS3Version
	object["capture_start_time"] = startTime.Format(time.RFC3339)
	object["capture_end_time"] = endTime.Format(time.RFC3339)
	object["number_of_flow_logs"] = len(flows)
	object["flow_logs"] = flows

	return object
}

func (s *encodeS3) Update(_ config.StageParam) {
	log.Warn("Encode S3 Writer, update not supported")
}

func (s *encodeS3) createObjectTimeoutLoop() {
	log.Debugf("entering createObjectTimeoutLoop")
	ticker := time.NewTicker(s.s3Params.WriteTimeout.Duration)
	for {
		select {
		case <-s.exitChan:
			log.Debugf("exiting createObjectTimeoutLoop because of signal")
			return
		case <-ticker.C:
			now := time.Now()
			log.Debugf("time now = %v, expiryTime = %v", now, s.expiryTime)
			s.mutex.Lock()
			_ = s.writeObject()
			s.mutex.Unlock()
		}
	}
}

// Encode queues entries to be sent to object store
func (s *encodeS3) Encode(entry config.GenericMap) {
	log.Debugf("Encode S3, entry = %v", entry)
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.pendingEntries = append(s.pendingEntries, entry)
	s.recordsWritten.Inc()
	if len(s.pendingEntries) >= s.s3Params.BatchSize {
		_ = s.writeObject()
	}
}

// NewEncodeS3 creates a new writer to S3
func NewEncodeS3(opMetrics *operational.Metrics, params config.StageParam) (Encoder, error) {
	configParams := api.EncodeS3{}
	if params.Encode != nil && params.Encode.S3 != nil {
		configParams = *params.Encode.S3
	}
	log.Debugf("NewEncodeS3, config = %v", configParams)
	s3Writer := &encodeS3Writer{
		s3Params: &configParams,
	}
	if configParams.WriteTimeout.Duration == time.Duration(0) {
		configParams.WriteTimeout = defaultTimeOut
	}
	if configParams.BatchSize == 0 {
		configParams.BatchSize = defaultBatchSize
	}

	s := &encodeS3{
		s3Params:          configParams,
		s3Writer:          s3Writer,
		recordsWritten:    opMetrics.CreateRecordsWrittenCounter(params.Name),
		pendingEntries:    make([]config.GenericMap, 0),
		expiryTime:        time.Now().Add(configParams.WriteTimeout.Duration),
		exitChan:          utils.ExitChannel(),
		streamID:          time.Now().Format(time.RFC3339),
		intervalStartTime: time.Now(),
		mutex:             &sync.Mutex{},
	}
	go s.createObjectTimeoutLoop()
	return s, nil
}

func (e *encodeS3Writer) connectS3(config *api.EncodeS3) (*minio.Client, error) {
	// Initialize s3 client object.
	minioOptions := minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKeyID, config.SecretAccessKey, ""),
		Secure: config.Secure,
	}
	s3Client, err := minio.New(config.Endpoint, &minioOptions)
	if err != nil {
		log.Errorf("Error when creating S3 client: %v", err)
		return nil, err
	}

	found, err := s3Client.BucketExists(context.Background(), config.Bucket)
	if err != nil {
		log.Errorf("Error accessing S3 bucket: %v", err)
		return nil, err
	}
	if found {
		log.Infof("S3 Bucket %s found", config.Bucket)
	}
	log.Debugf("s3Client = %#v", s3Client) // s3Client is now setup
	return s3Client, nil
}

func (e *encodeS3Writer) putObject(bucket string, objectName string, object map[string]interface{}) error {
	if e.s3Client == nil {
		s3Client, err := e.connectS3(e.s3Params)
		if s3Client == nil {
			return err
		}
		e.s3Client = s3Client
	}
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(object)
	if err != nil {
		log.Errorf("error encoding object: %v", err)
		return err
	}
	log.Debugf("encoded object = %v", b)
	// TBD: add necessary headers such as authorization (token), gzip, md5, etc
	uploadInfo, err := e.s3Client.PutObject(context.Background(), bucket, objectName, b, int64(b.Len()), minio.PutObjectOptions{ContentType: "application/octet-stream"})
	log.Debugf("uploadInfo = %v", uploadInfo)
	return err
}

func (e *encodeS3Writer) Update(_ config.StageParam) {
	log.Warn("Encode S3 Writer, update not supported")
}
