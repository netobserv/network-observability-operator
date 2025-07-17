/*
 * Copyright (C) 2022 IBM, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package utils

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
)

var floatType = reflect.TypeOf(float64(0))
var intType = reflect.TypeOf(int(0))
var uint32Type = reflect.TypeOf(uint32(0))
var uint64Type = reflect.TypeOf(uint64(0))
var int64Type = reflect.TypeOf(int64(0))
var stringType = reflect.TypeOf("")

// ConvertToFloat64 converts an unknown type to float
// Based on https://stackoverflow.com/a/20767884/2749989
func ConvertToFloat64(unk interface{}) (float64, error) {
	switch i := unk.(type) {
	case float64:
		return i, nil
	case float32:
		return float64(i), nil
	case int64:
		return float64(i), nil
	case int32:
		return float64(i), nil
	case int:
		return float64(i), nil
	case uint64:
		return float64(i), nil
	case uint32:
		return float64(i), nil
	case uint:
		return float64(i), nil
	case string:
		return strconv.ParseFloat(i, 64)
	default:
		v := reflect.ValueOf(unk)
		v = reflect.Indirect(v)
		if v.Type().ConvertibleTo(floatType) {
			fv := v.Convert(floatType)
			return fv.Float(), nil
		} else if v.Type().ConvertibleTo(stringType) {
			sv := v.Convert(stringType)
			s := sv.String()
			return strconv.ParseFloat(s, 64)
		} else {
			return math.NaN(), fmt.Errorf("can't convert %v to float64", v.Type())
		}
	}
}

func ConvertToUint32(unk interface{}) (uint32, error) {
	switch i := unk.(type) {
	case float64:
		return uint32(i), nil
	case float32:
		return uint32(i), nil
	case int64:
		return uint32(i), nil
	case int32:
		return uint32(i), nil
	case int:
		return uint32(i), nil
	case uint64:
		return uint32(i), nil
	case uint32:
		return i, nil
	case uint:
		return uint32(i), nil
	case string:
		res, err := strconv.ParseUint(i, 10, 32)
		return uint32(res), err
	default:
		v := reflect.ValueOf(unk)
		v = reflect.Indirect(v)
		if v.Type().ConvertibleTo(uint32Type) {
			fv := v.Convert(uint32Type)
			return uint32(fv.Uint()), nil
		} else if v.Type().ConvertibleTo(stringType) {
			sv := v.Convert(stringType)
			s := sv.String()
			res, err := strconv.ParseUint(s, 10, 32)
			return uint32(res), err
		} else {
			return 0, fmt.Errorf("can't convert %v to uint32", v.Type())
		}
	}
}

func ConvertToUint64(unk interface{}) (uint64, error) {
	switch i := unk.(type) {
	case float64:
		return uint64(i), nil
	case float32:
		return uint64(i), nil
	case int64:
		return uint64(i), nil
	case int32:
		return uint64(i), nil
	case int:
		return uint64(i), nil
	case uint64:
		return i, nil
	case uint32:
		return uint64(i), nil
	case uint:
		return uint64(i), nil
	case string:
		return strconv.ParseUint(i, 10, 64)
	default:
		v := reflect.ValueOf(unk)
		v = reflect.Indirect(v)
		if v.Type().ConvertibleTo(uint64Type) {
			fv := v.Convert(uint64Type)
			return fv.Uint(), nil
		} else if v.Type().ConvertibleTo(stringType) {
			sv := v.Convert(stringType)
			s := sv.String()
			return strconv.ParseUint(s, 10, 64)
		} else {
			return 0, fmt.Errorf("can't convert %v to uint64", v.Type())
		}
	}
}

func ConvertToInt64(unk interface{}) (int64, error) {
	switch i := unk.(type) {
	case float64:
		return int64(i), nil
	case float32:
		return int64(i), nil
	case int64:
		return i, nil
	case int32:
		return int64(i), nil
	case int:
		return int64(i), nil
	case uint64:
		return int64(i), nil
	case uint32:
		return int64(i), nil
	case uint:
		return int64(i), nil
	case string:
		return strconv.ParseInt(i, 10, 64)
	default:
		v := reflect.ValueOf(unk)
		v = reflect.Indirect(v)
		if v.Type().ConvertibleTo(int64Type) {
			fv := v.Convert(int64Type)
			return fv.Int(), nil
		} else if v.Type().ConvertibleTo(stringType) {
			sv := v.Convert(stringType)
			s := sv.String()
			return strconv.ParseInt(s, 10, 64)
		} else {
			return 0, fmt.Errorf("can't convert %v to int64", v.Type())
		}
	}
}

func ConvertToInt(unk interface{}) (int, error) {
	switch i := unk.(type) {
	case float64:
		return int(i), nil
	case float32:
		return int(i), nil
	case int64:
		return int(i), nil
	case int32:
		return int(i), nil
	case int:
		return i, nil
	case uint64:
		return int(i), nil
	case uint32:
		return int(i), nil
	case uint:
		return int(i), nil
	case string:
		res, err := strconv.ParseInt(i, 10, 64)
		return int(res), err
	default:
		v := reflect.ValueOf(unk)
		v = reflect.Indirect(v)
		if v.Type().ConvertibleTo(intType) {
			fv := v.Convert(intType)
			return int(fv.Int()), nil
		} else if v.Type().ConvertibleTo(stringType) {
			sv := v.Convert(stringType)
			s := sv.String()
			res, err := strconv.ParseInt(s, 10, 64)
			return int(res), err
		} else {
			return 0, fmt.Errorf("can't convert %v to int", v.Type())
		}
	}
}

func ConvertToUint(unk interface{}) (uint, error) {
	switch i := unk.(type) {
	case uint64:
		return uint(i), nil
	case uint32:
		return uint(i), nil
	case uint16:
		return uint(i), nil
	case uint:
		return uint(i), nil
	default:
		return 0, fmt.Errorf("can't convert %v to uint", i)
	}
}

func ConvertToBool(unk interface{}) (bool, error) {
	switch i := unk.(type) {
	case string:
		return strconv.ParseBool(i)
	case bool:
		return i, nil
	default:
		v := reflect.ValueOf(unk)
		v = reflect.Indirect(v)
		if v.Type().ConvertibleTo(intType) {
			sv := v.Convert(intType)
			s := sv.Int()
			switch s {
			case 0:
				return false, nil
			case 1:
				return true, nil
			default:
				return false, fmt.Errorf("can't convert %v (%v) to bool", s, v.Type())
			}
		} else {
			return false, fmt.Errorf("can't convert %v (%v) to bool", unk, v.Type())
		}
	}
}

func ConvertToString(unk interface{}) string {
	switch i := unk.(type) {
	case float64:
		return strconv.FormatFloat(i, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(i), 'f', -1, 32)
	case int64:
		return strconv.FormatInt(i, 10)
	case int32:
		return strconv.FormatInt(int64(i), 10)
	case int:
		return strconv.FormatInt(int64(i), 10)
	case uint64:
		return strconv.FormatUint(i, 10)
	case uint32:
		return strconv.FormatUint(uint64(i), 10)
	case uint:
		return strconv.FormatUint(uint64(i), 10)
	case string:
		return i
	default:
		return fmt.Sprintf("%v", unk)
	}
}
