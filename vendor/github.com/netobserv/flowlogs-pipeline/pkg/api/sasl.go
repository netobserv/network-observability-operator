package api

type SASLConfig struct {
	Type             string
	ClientIDPath     string `yaml:"clientIDPath,omitempty" json:"clientIDPath,omitempty" doc:"path to the client ID / SASL username"`
	ClientSecretPath string `yaml:"clientSecretPath,omitempty" json:"clientSecretPath,omitempty" doc:"path to the client secret / SASL password"`
}

type SASLTypeEnum struct {
	Plain       string `yaml:"plain" json:"plain" doc:"Plain SASL"`
	ScramSHA512 string `yaml:"scramSHA512" json:"scramSHA512" doc:"SCRAM/SHA512 SASL"`
}

func SASLTypeName(operation string) string {
	return GetEnumName(SASLTypeEnum{}, operation)
}
