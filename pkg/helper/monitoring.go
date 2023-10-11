package helper

import (
	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta2"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
)

func GetSecretOrConfigMap(file *flowslatest.FileReference) monitoringv1.SecretOrConfigMap {
	if file.Type == flowslatest.RefTypeConfigMap {
		return monitoringv1.SecretOrConfigMap{
			ConfigMap: &corev1.ConfigMapKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: file.Name,
				},
				Key: file.File,
			},
		}
	}
	return monitoringv1.SecretOrConfigMap{
		Secret: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: file.Name,
			},
			Key: file.File,
		},
	}
}
