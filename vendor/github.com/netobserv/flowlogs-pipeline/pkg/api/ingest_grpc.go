package api

type IngestGRPCProto struct {
	Port      int `yaml:"port" json:"port" doc:"the port number to listen on"`
	BufferLen int `yaml:"bufferLength" json:"bufferLength" doc:"the length of the ingest channel buffer, in groups of flows, containing each group hundreds of flows (default: 100)"`
}
