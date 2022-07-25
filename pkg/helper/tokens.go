package helper

import (
	corev1 "k8s.io/api/core/v1"
)

const TokensPath = "/var/run/secrets/tokens/"

// AppendTokenVolume will add a volume + volume mount for a service account token if defined
func AppendTokenVolume(volumes []corev1.Volume, volumeMounts []corev1.VolumeMount, name string, fileName string) ([]corev1.Volume, []corev1.VolumeMount) {
	volOut := append(volumes,
		corev1.Volume{
			Name: name,
			VolumeSource: corev1.VolumeSource{
				Projected: &corev1.ProjectedVolumeSource{
					Sources: []corev1.VolumeProjection{
						{
							ServiceAccountToken: &corev1.ServiceAccountTokenProjection{
								Path: fileName,
							},
						},
					},
				},
			},
		})
	vmOut := append(volumeMounts,
		corev1.VolumeMount{
			MountPath: TokensPath,
			Name:      name,
		},
	)
	return volOut, vmOut
}
