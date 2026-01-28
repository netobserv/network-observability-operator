//nolint:revive
package api

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Settings struct {
	Timeout metav1.Duration `json:"timeout,omitempty"`
}

func TestEncodeDuration(t *testing.T) {
	assert := assert.New(t)
	msgEnc, err := json.Marshal(&Settings{
		Timeout: metav1.Duration{Duration: time.Second * 2},
	})
	assert.Nil(err)
	assert.Equal(`{"timeout":"2s"}`, string(msgEnc))
}

func TestDecodeDurationAsString(t *testing.T) {
	assert := assert.New(t)
	var dec Settings
	err := json.Unmarshal([]byte(`{"timeout":"2s"}`), &dec)
	assert.Nil(err)
	assert.Equal(metav1.Duration{Duration: time.Second * 2}, dec.Timeout)
}
