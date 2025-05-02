package filters

import (
	"regexp"
	"strings"

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

func NumEquals(key string, filterValue int) Predicate {
	return castIntAndCheck(key, func(i int) bool { return i == filterValue })
}

func NumNotEquals(key string, filterValue int) Predicate {
	return castIntAndCheck(key, func(i int) bool { return i != filterValue })
}

func LessThan(key string, filterValue int) Predicate {
	return castIntAndCheck(key, func(i int) bool { return i < filterValue })
}

func GreaterThan(key string, filterValue int) Predicate {
	return castIntAndCheck(key, func(i int) bool { return i > filterValue })
}

func LessOrEqualThan(key string, filterValue int) Predicate {
	return castIntAndCheck(key, func(i int) bool { return i <= filterValue })
}

func GreaterOrEqualThan(key string, filterValue int) Predicate {
	return castIntAndCheck(key, func(i int) bool { return i >= filterValue })
}

func castIntAndCheck(key string, check func(int) bool) Predicate {
	return func(flow config.GenericMap) bool {
		if val, found := flow[key]; found {
			if cast, err := utils.ConvertToInt(val); err == nil {
				return check(cast)
			}
		}
		return false
	}
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
