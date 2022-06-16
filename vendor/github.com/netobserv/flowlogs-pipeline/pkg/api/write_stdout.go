package api

type WriteStdout struct {
	Format string `yaml:"format,omitempty" json:"format,omitempty" doc:"the format of each line: printf (default) or json"`
}
