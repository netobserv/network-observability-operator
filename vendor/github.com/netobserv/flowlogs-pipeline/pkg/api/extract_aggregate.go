package api

type AggregateBy []string
type AggregateOperation string

type AggregateDefinition struct {
	Name      string             `yaml:"name,omitempty" json:"name,omitempty" doc:"description of aggregation result"`
	By        AggregateBy        `yaml:"by,omitempty" json:"by,omitempty" doc:"list of fields on which to aggregate"`
	Operation AggregateOperation `yaml:"operation,omitempty" json:"operation,omitempty" doc:"sum, min, max, avg or raw_values"`
	RecordKey string             `yaml:"recordKey,omitempty" json:"recordKey,omitempty" doc:"internal field on which to perform the operation"`
	TopK      int                `yaml:"topK,omitempty" json:"topK,omitempty" doc:"number of highest incidence to report (default - report all)"`
}
