package api

type IngestGRPCProto struct {
	Port         int    `yaml:"port,omitempty" json:"port,omitempty" doc:"the port number to listen on"`
	BufferLen    int    `yaml:"bufferLength,omitempty" json:"bufferLength,omitempty" doc:"the length of the ingest channel buffer, in groups of flows, containing each group hundreds of flows (default: 100)"`
	CertPath     string `yaml:"certPath,omitempty" json:"certPath,omitempty" doc:"path of the TLS certificate, if any"`
	KeyPath      string `yaml:"keyPath,omitempty" json:"keyPath,omitempty" doc:"path of the TLS certificate key, if any"`
	ClientCAPath string `yaml:"clientCAPath,omitempty" json:"clientCAPath,omitempty" doc:"path of the client TLS CA, if any, for mutual TLS"`
}
