package v1beta2

// Validation warning messages
const (
	// MetricsIncludeListWarning is shown when both includeList and additionalIncludeList are set
	MetricsIncludeListWarning = "Both spec.processor.metrics.includeList and spec.processor.metrics.additionalIncludeList are set. " +
		"When includeList is set, it replaces the default metrics entirely, and additionalIncludeList is ignored. " +
		"Use includeList to override defaults, or use additionalIncludeList alone to append to defaults."

	// LokiS3Warning is shown when Loki and an S3 exporter are both enabled (not recommended).
	LokiS3Warning = "Loki and an S3 exporter are both enabled. Running two raw-flow stores is not recommended; prefer Prometheus + S3 (Loki disabled) or Loki alone."

	// LokiFlowBufferWarning is shown when Loki is on and flowBuffer is explicitly enabled.
	LokiFlowBufferWarning = "Loki is enabled and spec.processor.flowBuffer.enable is true. The flow buffer uses extra memory and is unused for Console raw queries while Loki serves raw flows."

	// LokiS3WarningSelectable is shown when Loki is on and consolePlugin.s3 is enabled.
	// S3 remains selectable in the Console; Auto still prefers Loki for raw flows.
	LokiS3WarningSelectable = "Loki is enabled and spec.consolePlugin.s3.enable is true. Auto/default raw queries prefer Loki; users can still select S3 explicitly in the Console."
)
