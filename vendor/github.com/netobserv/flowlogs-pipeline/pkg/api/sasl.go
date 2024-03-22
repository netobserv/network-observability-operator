package api

type SASLConfig struct {
	Type             SASLTypeEnum `yaml:"type,omitempty" json:"type,omitempty" doc:"SASL type"`
	ClientIDPath     string       `yaml:"clientIDPath,omitempty" json:"clientIDPath,omitempty" doc:"path to the client ID / SASL username"`
	ClientSecretPath string       `yaml:"clientSecretPath,omitempty" json:"clientSecretPath,omitempty" doc:"path to the client secret / SASL password"`
}

type SASLTypeEnum string

const (
	// For doc generation, enum definitions must match format `Constant Type = "value" // doc`
	SASLPlain       SASLTypeEnum = "plain"       // Plain SASL
	SASLScramSHA512 SASLTypeEnum = "scramSHA512" // SCRAM/SHA512 SASL
)
