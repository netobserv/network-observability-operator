package flp

import "embed"

//go:embed metrics_definitions
var FlpMetricsConfig embed.FS

var FlpMetricsConfigDir = "metrics_definitions"
