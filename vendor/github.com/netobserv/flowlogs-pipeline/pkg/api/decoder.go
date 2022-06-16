package api

type Decoder struct {
	Type string `yaml:"type" json:"type" enum:"DecoderEnum" doc:"one of the following:"`
}

type DecoderEnum struct {
	JSON     string `yaml:"json" json:"json" doc:"JSON decoder"`
	Protobuf string `yaml:"protobuf" json:"protobuf" doc:"Protobuf decoder"`
}

func DecoderName(decoder string) string {
	return GetEnumName(DecoderEnum{}, decoder)
}
