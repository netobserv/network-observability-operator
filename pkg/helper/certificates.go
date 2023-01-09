package helper

import (
	"fmt"

	"github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	corev1 "k8s.io/api/core/v1"
)

type AddCertCallback = func(string, *v1alpha1.CertificateReference)

// AppendCertVolumes will add a volume + volume mount for a CA cert if defined, and another volume + volume mount for a user cert if defined.
// It does nothing if neither is defined.
func AppendCertVolumes(volumes []corev1.Volume, volumeMounts []corev1.VolumeMount, config *v1alpha1.ClientTLS, name string, onCertAdded AddCertCallback) ([]corev1.Volume, []corev1.VolumeMount) {
	volOut := volumes
	vmOut := volumeMounts
	if config.CACert.Name != "" {
		vol, vm := buildVolume(&config.CACert, constants.CertCAName(name), onCertAdded)
		volOut = append(volOut, vol)
		vmOut = append(vmOut, vm)
	}
	if config.UserCert.Name != "" {
		vol, vm := buildVolume(&config.UserCert, constants.CertUserName(name), onCertAdded)
		volOut = append(volOut, vol)
		vmOut = append(vmOut, vm)
	}
	return volOut, vmOut
}

func AppendSingleCertVolumes(volumes []corev1.Volume, volumeMounts []corev1.VolumeMount, config *v1alpha1.CertificateReference, name string, onCertAdded AddCertCallback) ([]corev1.Volume, []corev1.VolumeMount) {
	volOut := volumes
	vmOut := volumeMounts
	if config.Name != "" {
		vol, vm := buildVolume(config, name, onCertAdded)
		volOut = append(volOut, vol)
		vmOut = append(vmOut, vm)
	}
	return volOut, vmOut
}

func buildVolume(ref *v1alpha1.CertificateReference, name string, onCertAdded AddCertCallback) (corev1.Volume, corev1.VolumeMount) {
	var vol corev1.Volume
	if ref.Type == v1alpha1.CertRefTypeConfigMap {
		vol = corev1.Volume{
			Name: name,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: ref.Name,
					},
				},
			},
		}
	} else {
		vol = corev1.Volume{
			Name: name,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: ref.Name,
				},
			},
		}
	}
	onCertAdded(name, ref)
	return vol, corev1.VolumeMount{
		Name:      name,
		ReadOnly:  true,
		MountPath: "/var/" + name,
	}
}

func getPath(base, suffix, file string) string {
	if len(suffix) > 0 {
		return fmt.Sprintf("/var/%s-%s/%s", base, suffix, file)
	}
	return fmt.Sprintf("/var/%s/%s", base, file)
}

// GetCACertPath returns the CA cert path that corresponds to a volume/volume mount created with "AppendCertVolumes"
// When not available, an empty string is returned.
func GetCACertPath(config *v1alpha1.ClientTLS, name string) string {
	if config.CACert.Name != "" {
		return getPath(name, constants.CertCASuffix, config.CACert.CertFile)
	}
	return ""
}

// GetUserCertPath returns the user cert path that corresponds to a volume/volume mount created with "AppendCertVolumes"
// When not available, an empty string is returned.
func GetUserCertPath(config *v1alpha1.ClientTLS, name string) string {
	if config.UserCert.Name != "" {
		return getPath(name, constants.CertUserSuffix, config.UserCert.CertFile)
	}
	return ""
}

// GetUserKeyPath returns the user private key path that corresponds to a volume/volume mount created with "AppendCertVolumes"
// When not available, an empty string is returned.
func GetUserKeyPath(config *v1alpha1.ClientTLS, name string) string {
	if config.UserCert.Name != "" {
		return getPath(name, constants.CertUserSuffix, config.UserCert.CertKey)
	}
	return ""
}

func GetSingleCertPath(config *v1alpha1.CertificateReference, name string) string {
	if config.Name != "" {
		return getPath(name, "", config.CertFile)
	}
	return ""
}

func GetSingleKeyPath(config *v1alpha1.CertificateReference, name string) string {
	if config.Name != "" {
		return getPath(name, "", config.CertKey)
	}
	return ""
}
