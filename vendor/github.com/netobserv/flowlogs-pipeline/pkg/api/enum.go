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

package api

import (
	"log"
	"reflect"
)

type enums struct {
	PromEncodeOperationEnum       PromEncodeOperationEnum
	TransformNetworkOperationEnum TransformNetworkOperationEnum
	TransformFilterOperationEnum  TransformFilterOperationEnum
	TransformGenericOperationEnum TransformGenericOperationEnum
	KafkaEncodeBalancerEnum       KafkaEncodeBalancerEnum
	SASLTypeEnum                  SASLTypeEnum
	ConnTrackOperationEnum        ConnTrackOperationEnum
	ConnTrackOutputRecordTypeEnum ConnTrackOutputRecordTypeEnum
	DecoderEnum                   DecoderEnum
	FilterOperationEnum           FilterOperationEnum
}

type enumNameCacheKey struct {
	enum      interface{}
	operation string
}

var enumNamesCache = map[enumNameCacheKey]string{}

func init() {
	populateEnumCache()
}

func populateEnumCache() {
	enumStruct := enums{}
	e := reflect.ValueOf(&enumStruct).Elem()
	for i := 0; i < e.NumField(); i++ {
		eType := e.Type().Field(i).Type
		eValue := e.Field(i).Interface()
		for j := 0; j < eType.NumField(); j++ {
			fName := eType.Field(j).Name
			key := enumNameCacheKey{enum: eValue, operation: fName}
			d := reflect.ValueOf(eValue)
			field, _ := d.Type().FieldByName(fName)
			tag := field.Tag.Get(TagYaml)
			enumNamesCache[key] = tag
		}
	}
}

// GetEnumName gets the name of an enum value from the representing enum struct based on `TagYaml` tag.
func GetEnumName(enum interface{}, operation string) string {
	key := enumNameCacheKey{enum: enum, operation: operation}
	cachedValue, found := enumNamesCache[key]
	if found {
		return cachedValue
	} else {
		log.Panicf("can't find name '%s' in enum %v", operation, enum)
		return ""
	}
}

// GetEnumReflectionTypeByFieldName gets the enum struct `reflection Type` from the name of the struct (using fields from `enums{}` struct).
func GetEnumReflectionTypeByFieldName(enumName string) reflect.Type {
	d := reflect.ValueOf(enums{})
	field, found := d.Type().FieldByName(enumName)
	if !found {
		log.Panicf("can't find enumName %s in enums", enumName)
		return nil
	}

	return field.Type
}
