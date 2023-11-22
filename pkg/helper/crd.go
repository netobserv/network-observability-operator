package helper

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	crd                   *apiextensionsv1.CustomResourceDefinition
	quoteRegex            = regexp.MustCompile(`^"(.*)"$`)
	AgentAdvancedPath     = []string{"spec", "agent", "advanced"}
	ProcessorAdvancedPath = []string{"spec", "processor", "advanced"}
	PluginAdvancedPath    = []string{"spec", "consolePlugin", "advanced"}
	LokiAdvancedPath      = []string{"spec", "loki", "advanced"}
)

func ParseCRD(bytes []byte) error {
	crdScheme := runtime.NewScheme()
	err := apiextensionsv1.AddToScheme(crdScheme)
	if err != nil {
		return err
	}
	err = apiextensionsv1beta1.AddToScheme(crdScheme)
	if err != nil {
		return err
	}
	crdCodecFactory := serializer.NewCodecFactory(crdScheme)
	crdDeserializer := crdCodecFactory.UniversalDeserializer()
	crdObj, _, err := crdDeserializer.Decode(bytes, nil, &apiextensionsv1.CustomResourceDefinition{})
	if err != nil {
		return err
	}
	crd = crdObj.(*apiextensionsv1.CustomResourceDefinition)

	return nil
}

func SetCRD(v *apiextensionsv1.CustomResourceDefinition) {
	crd = v
}

func GetAdvancedDurationValue(path []string, field string, value *v1.Duration) *v1.Duration {
	if value != nil && !IsDefaultValue(path, field, value.Duration.String()) {
		return value
	}
	return nil
}

func GetAdvancedBoolValue(path []string, field string, value *bool) *bool {
	if value != nil && !IsDefaultValue(path, field, *value) {
		return value
	}
	return nil
}

func GetAdvancedInt32Value(path []string, field string, value *int32) *int32 {
	if value != nil && !IsDefaultValue(path, field, *value) {
		return value
	}
	return nil
}

func GetAdvancedMapValue(path []string, field string, value *map[string]string) *map[string]string {
	bytes, _ := json.Marshal(value)
	if !IsDefaultValue(path, field, string(bytes)) {
		return value
	}
	return nil
}

func IsDefaultValue(path []string, field string, value interface{}) bool {
	defaultValueStr := GetFieldDefaultString(path, field)
	switch value.(type) {
	case string:
		return value == defaultValueStr
	case bool:
		return fmt.Sprintf("%t", value) == defaultValueStr
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", value) == defaultValueStr
	case float32, float64:
		return fmt.Sprintf("%f", value) == defaultValueStr
	default:
		return false
	}
}

func GetFieldDefaultString(path []string, field string) string {
	bytes := GetFieldDefault(path, field)
	if len(bytes) > 0 {
		return quoteRegex.ReplaceAllString(string(bytes), `$1`)
	}
	return ""
}

func GetFieldDefaultInt32(path []string, field string) int32 {
	defaultValueStr := GetFieldDefaultString(path, field)
	intVar, _ := strconv.ParseInt(defaultValueStr, 0, 32)
	return int32(intVar)
}

func GetFieldDefaultBool(path []string, field string) bool {
	defaultValueStr := GetFieldDefaultString(path, field)
	return defaultValueStr == "true"
}

func GetFieldDefaultDuration(path []string, field string) v1.Duration {
	duration := v1.Duration{}
	_ = duration.UnmarshalJSON(GetFieldDefault(path, field))
	return duration
}

func GetFieldDefaultMapString(path []string, field string) map[string]string {
	bytes := GetFieldDefault(path, field)
	if len(bytes) > 0 {
		m := make(map[string]string)
		_ = json.Unmarshal(bytes, &m)
		return m
	}
	return map[string]string{}
}

func GetFieldDefault(path []string, field string) []byte {
	pathProperties := getPathProperties(path)
	if fieldSchema, ok := pathProperties[field]; ok {
		return fieldSchema.Default.Raw
	}
	return []byte{}
}

func GetValueOrDefaultInt32(path []string, field string, value *int32) int32 {
	if value != nil {
		return *value
	}
	return GetFieldDefaultInt32(path, field)
}

func GetValueOrDefaultMapString(path []string, field string, value *map[string]string) map[string]string {
	if value != nil {
		return *value
	}
	return GetFieldDefaultMapString(path, field)
}

func getPathProperties(path []string) map[string]apiextensionsv1.JSONSchemaProps {
	schema := getSchema()
	if schema == nil {
		return map[string]apiextensionsv1.JSONSchemaProps{}
	}
	properties := schema.Properties
	for _, key := range path {
		if val, ok := properties[key]; ok {
			properties = val.Properties
		}
	}
	return properties
}

func getSchema() *apiextensionsv1.JSONSchemaProps {
	if crd == nil {
		return nil
	}
	versions := crd.Spec.Versions
	if len(versions) > 0 {
		lastVersion := versions[len(versions)-1]
		return lastVersion.Schema.OpenAPIV3Schema
	}
	return nil
}
