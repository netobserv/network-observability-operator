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

func (w *Watcher) refFromConfigOrSecret(cos *flowslatest.ConfigOrSecret, keys []string) objectRef {
	ns := cos.Namespace
	if ns == "" {
		ns = w.defaultNamespace
	}
	return objectRef{
		kind:      cos.Type,
		name:      cos.Name,
		namespace: ns,
		keys:      keys,
	}
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
