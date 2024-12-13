package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"
)

func SetupSASLMechanism(cfg *api.SASLConfig) (sasl.Mechanism, error) {
	// Read client ID
	id, err := os.ReadFile(cfg.ClientIDPath)
	if err != nil {
		return nil, err
	}
	strID := strings.TrimSpace(string(id))
	// Read password
	pwd, err := os.ReadFile(cfg.ClientSecretPath)
	if err != nil {
		return nil, err
	}
	strPwd := strings.TrimSpace(string(pwd))
	var mechanism sasl.Mechanism
	switch cfg.Type {
	case api.SASLPlain:
		mechanism = plain.Mechanism{Username: strID, Password: strPwd}
	case api.SASLScramSHA512:
		mechanism, err = scram.Mechanism(scram.SHA512, strID, strPwd)
	default:
		return nil, fmt.Errorf("unknown SASL type: %s", cfg.Type)
	}
	if err != nil {
		return nil, err
	}
	return mechanism, nil
}
