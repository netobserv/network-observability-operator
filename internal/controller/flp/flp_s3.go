/*
Copyright 2026.

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

package flp

import (
	"context"
	"fmt"

	flowslatest "github.com/netobserv/netobserv-operator/api/flowcollector/v1beta2"
	"github.com/netobserv/netobserv-operator/internal/controller/reconcilers"
	"github.com/netobserv/netobserv-operator/internal/pkg/watchers"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	s3AccessKeyIDKey     = "accessKeyId"
	s3SecretAccessKeyKey = "secretAccessKey"
)

// loadS3ExporterCredentials reads S3 credential Secrets for FLP EncodeS3
// (plaintext accessKeyId / secretAccessKey — FLP does not support file paths).
// Also watches Secrets so pods restart on rotation via config digest / annotations.
func loadS3ExporterCredentials(ctx context.Context, info *reconcilers.Instance, desired *flowslatest.FlowCollectorSpec) (map[string]s3Credentials, error) {
	out := map[string]s3Credentials{}
	if info == nil || desired == nil {
		return out, nil
	}
	for i, exporter := range desired.Exporters {
		if exporter == nil || exporter.Type != flowslatest.S3Exporter {
			continue
		}
		name := fmt.Sprintf("s3-export-%d", i)
		secretName := exporter.S3.Credentials.Name
		if secretName == "" {
			return nil, fmt.Errorf("exporters[%d] S3 credentials.name is empty", i)
		}
		ns := desired.GetNamespace()
		if info.Watcher != nil {
			ref := flowslatest.FileReference{
				Type: flowslatest.RefTypeSecret,
				Name: secretName,
				File: s3AccessKeyIDKey,
			}
			if _, err := info.Watcher.ProcessFileReference(ctx, info.Client, ref, ns); err != nil {
				return nil, fmt.Errorf("watching S3 credentials secret %s: %w", secretName, err)
			}
		}
		var secret corev1.Secret
		if err := info.Client.Get(ctx, types.NamespacedName{Name: secretName, Namespace: ns}, &secret); err != nil {
			return nil, fmt.Errorf("getting S3 credentials secret %s/%s: %w", ns, secretName, err)
		}
		accessKey := string(secret.Data[s3AccessKeyIDKey])
		secretKey := string(secret.Data[s3SecretAccessKeyKey])
		if accessKey == "" || secretKey == "" {
			return nil, fmt.Errorf("S3 credentials secret %s must contain keys %q and %q", secretName, s3AccessKeyIDKey, s3SecretAccessKeyKey)
		}
		out[name] = s3Credentials{accessKeyID: accessKey, secretAccessKey: secretKey}
	}
	return out, nil
}

// annotateS3ExporterSecrets adds digest annotations so FLP pods restart when S3 secrets rotate.
func annotateS3ExporterSecrets(ctx context.Context, info *reconcilers.Common, exp []*flowslatest.FlowCollectorExporter, annotations map[string]string) error {
	for i, exporter := range exp {
		if exporter == nil || exporter.Type != flowslatest.S3Exporter {
			continue
		}
		ref := flowslatest.FileReference{
			Type: flowslatest.RefTypeSecret,
			Name: exporter.S3.Credentials.Name,
			File: s3AccessKeyIDKey,
		}
		digest, err := info.Watcher.ProcessFileReference(ctx, info.Client, ref, info.Namespace)
		if err != nil {
			return err
		}
		if digest != "" {
			annotations[watchers.Annotation(fmt.Sprintf("s3-export-%d", i))] = digest
		}
	}
	return nil
}
