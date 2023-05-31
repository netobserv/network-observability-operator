package watchers

import (
	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta1"
)

type objectRef struct {
	kind      flowslatest.MountableType
	name      string
	namespace string
	keys      []string
}

type ConfigOrSecret struct {
	Type      flowslatest.MountableType
	Name      string
	Namespace string
}

func (w *Watcher) refFromCert(cert *flowslatest.CertificateReference) objectRef {
	ns := cert.Namespace
	if ns == "" {
		ns = w.defaultNamespace
	}
	keys := []string{cert.CertFile}
	if cert.CertKey != "" {
		keys = append(keys, cert.CertKey)
	}
	return objectRef{
		kind:      cert.Type,
		name:      cert.Name,
		namespace: ns,
		keys:      keys,
	}
}
