package api

type AggregateBy []string
type AggregateOperation string

type AggregateDefinition struct {
	Name      string             `yaml:"name" json:"name" doc:"description of aggregation result"`
	By        AggregateBy        `yaml:"by" json:"by" doc:"list of fields on which to aggregate"`
	Operation AggregateOperation `yaml:"operation" json:"operation" doc:"sum, min, max, avg or raw_values"`
	RecordKey string             `yaml:"recordKey" json:"recordKey" doc:"internal field on which to perform the operation"`
	TopK      int                `yaml:"topK" json:"topK" doc:"number of highest incidence to report (default - report all)"`
}
