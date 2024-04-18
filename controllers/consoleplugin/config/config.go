package config

import (
	_ "embed"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
)

type ServerConfig struct {
	Port        int    `yaml:"port,omitempty" json:"port,omitempty"`
	MetricsPort int    `yaml:"metricsPort,omitempty" json:"metricsPort,omitempty"`
	CertPath    string `yaml:"certPath,omitempty" json:"certPath,omitempty"`
	KeyPath     string `yaml:"keyPath,omitempty" json:"keyPath,omitempty"`
	CORSOrigin  string `yaml:"corsOrigin,omitempty" json:"corsOrigin,omitempty"`
	CORSMethods string `yaml:"corsMethods,omitempty" json:"corsMethods,omitempty"`
	CORSHeaders string `yaml:"corsHeaders,omitempty" json:"corsHeaders,omitempty"`
	CORSMaxAge  string `yaml:"corsMaxAge,omitempty" json:"corsMaxAge,omitempty"`
}

type LokiConfig struct {
	URL    string   `yaml:"url" json:"url"`
	Labels []string `yaml:"labels" json:"labels"`

	StatusURL          string `yaml:"statusUrl,omitempty" json:"statusUrl,omitempty"`
	Timeout            string `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	TenantID           string `yaml:"tenantID,omitempty" json:"tenantID,omitempty"`
	TokenPath          string `yaml:"tokenPath,omitempty" json:"tokenPath,omitempty"`
	SkipTLS            bool   `yaml:"skipTls,omitempty" json:"skipTls,omitempty"`
	CAPath             string `yaml:"caPath,omitempty" json:"caPath,omitempty"`
	StatusSkipTLS      bool   `yaml:"statusSkipTls,omitempty" json:"statusSkipTls,omitempty"`
	StatusCAPath       string `yaml:"statusCaPath,omitempty" json:"statusCaPath,omitempty"`
	StatusUserCertPath string `yaml:"statusUserCertPath,omitempty" json:"statusUserCertPath,omitempty"`
	StatusUserKeyPath  string `yaml:"statusUserKeyPath,omitempty" json:"statusUserKeyPath,omitempty"`
	UseMocks           bool   `yaml:"useMocks,omitempty" json:"useMocks,omitempty"`
	ForwardUserToken   bool   `yaml:"forwardUserToken,omitempty" json:"forwardUserToken,omitempty"`
	AuthCheck          string `yaml:"authCheck,omitempty" json:"authCheck,omitempty"`
}

type ColumnConfig struct {
	ID   string `yaml:"id" json:"id"`
	Name string `yaml:"name" json:"name"`

	Group      string   `yaml:"group,omitempty" json:"group,omitempty"`
	Field      string   `yaml:"field,omitempty" json:"field,omitempty"`
	Fields     []string `yaml:"fields,omitempty" json:"fields,omitempty"`
	Calculated string   `yaml:"calculated,omitempty" json:"calculated,omitempty"`
	Tooltip    string   `yaml:"tooltip,omitempty" json:"tooltip,omitempty"`
	DocURL     string   `yaml:"docURL,omitempty" json:"docURL,omitempty"`
	Filter     string   `yaml:"filter,omitempty" json:"filter,omitempty"`
	Default    bool     `yaml:"default,omitempty" json:"default,omitempty"`
	Width      int      `yaml:"width,omitempty" json:"width,omitempty"`
	Feature    string   `yaml:"feature" json:"feature"`
}

type FilterConfig struct {
	ID        string `yaml:"id" json:"id"`
	Name      string `yaml:"name" json:"name"`
	Component string `yaml:"component" json:"component"`

	Category               string `yaml:"category,omitempty" json:"category,omitempty"`
	AutoCompleteAddsQuotes bool   `yaml:"autoCompleteAddsQuotes,omitempty" json:"autoCompleteAddsQuotes,omitempty"`
	Hint                   string `yaml:"hint,omitempty" json:"hint,omitempty"`
	Examples               string `yaml:"examples,omitempty" json:"examples,omitempty"`
	DocURL                 string `yaml:"docUrl,omitempty" json:"docUrl,omitempty"`
	Placeholder            string `yaml:"placeholder,omitempty" json:"placeholder,omitempty"`
}

type CardinalityWarn string

const (
	CardinalityWarnAvoid   CardinalityWarn = "avoid"
	CardinalityWarnCareful CardinalityWarn = "careful"
	CardinalityWarnFine    CardinalityWarn = "fine"
)

type FieldConfig struct {
	Name            string          `yaml:"name" json:"name"`
	Type            string          `yaml:"type" json:"type"`
	Description     string          `yaml:"description" json:"description"`
	LokiLabel       bool            `yaml:"lokiLabel,omitempty" json:"lokiLabel,omitempty"`
	Filter          string          `yaml:"filter,omitempty" json:"filter,omitempty"`
	CardinalityWarn CardinalityWarn `yaml:"cardinalityWarn,omitempty" json:"cardinalityWarn,omitempty"`
}

type Deduper struct {
	Mark  bool `yaml:"mark" json:"mark"`
	Merge bool `yaml:"merge" json:"merge"`
}

type FrontendConfig struct {
	RecordTypes []api.ConnTrackOutputRecordTypeEnum `yaml:"recordTypes" json:"recordTypes"`
	Columns     []ColumnConfig                      `yaml:"columns" json:"columns"`
	Sampling    int                                 `yaml:"sampling" json:"sampling"`
	Features    []string                            `yaml:"features" json:"features"`
	Deduper     Deduper                             `yaml:"deduper" json:"deduper"`
	Fields      []FieldConfig                       `yaml:"fields" json:"fields"`

	PortNaming      flowslatest.ConsolePluginPortConfig `yaml:"portNaming,omitempty" json:"portNaming,omitempty"`
	Filters         []FilterConfig                      `yaml:"filters,omitempty" json:"filters,omitempty"`
	QuickFilters    []flowslatest.QuickFilter           `yaml:"quickFilters,omitempty" json:"quickFilters,omitempty"`
	AlertNamespaces []string                            `yaml:"alertNamespaces,omitempty" json:"alertNamespaces,omitempty"`
}

type PluginConfig struct {
	Server   ServerConfig   `yaml:"server" json:"server"`
	Loki     LokiConfig     `yaml:"loki" json:"loki"`
	Frontend FrontendConfig `yaml:"frontend" json:"frontend"`
}

//go:embed static-frontend-config.yaml
var staticFrontendConfig []byte

func LoadStaticFrontendConfig() []byte {
	return staticFrontendConfig
}
