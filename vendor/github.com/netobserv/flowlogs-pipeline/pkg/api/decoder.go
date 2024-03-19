package api

type Decoder struct {
	Type DecoderEnum `yaml:"type" json:"type" doc:"(enum) one of the following:"`
}

type DecoderEnum string

const (
	// For doc generation, enum definitions must match format `Constant Type = "value" // doc`
	DecoderJSON     DecoderEnum = "json"     // JSON decoder
	DecoderProtobuf DecoderEnum = "protobuf" // Protobuf decoder
)
