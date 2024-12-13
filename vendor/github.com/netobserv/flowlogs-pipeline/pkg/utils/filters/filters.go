package filters

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/netobserv/flowlogs-pipeline/pkg/utils"
)

type Predicate func(config.GenericMap) bool

var variableExtractor = regexp.MustCompile(`\$\(([^\)]+)\)`)

func Presence(key string) Predicate {
	return func(flow config.GenericMap) bool {
		_, found := flow[key]
		return found
	}
}

func Absence(key string) Predicate {
	return func(flow config.GenericMap) bool {
		_, found := flow[key]
		return !found
	}
}

func Equal(key string, filterValue any, convertString bool) Predicate {
	varLookups := extractVarLookups(filterValue)
	if len(varLookups) > 0 {
		return func(flow config.GenericMap) bool {
			if val, found := flow[key]; found {
				// Variable injection => convert to string
				sVal, ok := val.(string)
				if !ok {
					sVal = utils.ConvertToString(val)
				}
				injected := injectVars(flow, filterValue.(string), varLookups)
				return sVal == injected
			}
			return false
		}
	}
	if convertString {
		return func(flow config.GenericMap) bool {
			if val, found := flow[key]; found {
				sVal, ok := val.(string)
				if !ok {
					sVal = utils.ConvertToString(val)
				}
				return sVal == filterValue
			}
			return false
		}
	}
	return func(flow config.GenericMap) bool {
		if val, found := flow[key]; found {
			return val == filterValue
		}
		return false
	}
}

func NotEqual(key string, filterValue any, convertString bool) Predicate {
	pred := Equal(key, filterValue, convertString)
	return func(flow config.GenericMap) bool { return !pred(flow) }
}

func Regex(key string, filterRegex *regexp.Regexp) Predicate {
	return func(flow config.GenericMap) bool {
		if val, found := flow[key]; found {
			sVal, ok := val.(string)
			if !ok {
				sVal = utils.ConvertToString(val)
			}
			return filterRegex.MatchString(sVal)
		}
		return false
	}
}

func NotRegex(key string, filterRegex *regexp.Regexp) Predicate {
	pred := Regex(key, filterRegex)
	return func(flow config.GenericMap) bool { return !pred(flow) }
}

func extractVarLookups(value any) [][]string {
	// Extract list of variables to lookup
	// E.g: filter "$(SrcAddr):$(SrcPort)" would return [SrcAddr,SrcPort]
	if sVal, isString := value.(string); isString {
		if len(sVal) > 0 {
			return variableExtractor.FindAllStringSubmatch(sVal, -1)
		}
	}
	return nil
}

func injectVars(flow config.GenericMap, filterValue string, varLookups [][]string) string {
	injected := filterValue
	for _, matchGroup := range varLookups {
		var value string
		if rawVal, found := flow[matchGroup[1]]; found {
			if sVal, ok := rawVal.(string); ok {
				value = sVal
			} else {
				value = utils.ConvertToString(rawVal)
			}
		}
		injected = strings.ReplaceAll(injected, matchGroup[0], value)
	}
	return injected
}

func FromKeepEntry(from *api.KeepEntryRule) (Predicate, error) {
	switch from.Type {
	case api.KeepEntryIfExists:
		return Presence(from.KeepEntry.Input), nil
	case api.KeepEntryIfDoesntExist:
		return Absence(from.KeepEntry.Input), nil
	case api.KeepEntryIfEqual:
		return Equal(from.KeepEntry.Input, from.KeepEntry.Value, true), nil
	case api.KeepEntryIfNotEqual:
		return NotEqual(from.KeepEntry.Input, from.KeepEntry.Value, true), nil
	case api.KeepEntryIfRegexMatch:
		if r, err := compileRegex(from.KeepEntry); err != nil {
			return nil, err
		} else {
			return Regex(from.KeepEntry.Input, r), nil
		}
	case api.KeepEntryIfNotRegexMatch:
		if r, err := compileRegex(from.KeepEntry); err != nil {
			return nil, err
		} else {
			return NotRegex(from.KeepEntry.Input, r), nil
		}
	}
	return nil, fmt.Errorf("keep entry rule type not recognized: %s", from.Type)
}

func compileRegex(from *api.TransformFilterGenericRule) (*regexp.Regexp, error) {
	s, ok := from.Value.(string)
	if !ok {
		return nil, fmt.Errorf("invalid regex keep rule: rule value must be a string [%v]", from)
	}
	r, err := regexp.Compile(s)
	if err != nil {
		return nil, fmt.Errorf("invalid regex keep rule: cannot compile regex [%w]", err)
	}
	return r, nil
}
