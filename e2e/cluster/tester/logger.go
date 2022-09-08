package tester

import (
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func InitLogger(keysAndValues ...interface{}) logr.Logger {
	opts := zap.Options{Development: true}
	return zap.New(zap.UseFlagOptions(&opts)).WithValues(keysAndValues...)
}
