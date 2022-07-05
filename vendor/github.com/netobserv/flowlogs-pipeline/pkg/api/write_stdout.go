package api

type WriteStdout struct {
	Format string `yaml:"format,omitempty" json:"format,omitempty" doc:"the format of each line: printf (default - writes using golang's default map printing), fields (writes one key and value field per line) or json"`
}
