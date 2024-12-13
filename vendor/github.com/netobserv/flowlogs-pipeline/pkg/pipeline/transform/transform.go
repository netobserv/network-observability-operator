/*
 * Copyright (C) 2021 IBM, Inc.
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

package transform

import (
	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/sirupsen/logrus"
)

type Transformer interface {
	Transform(in config.GenericMap) (config.GenericMap, bool)
}

type transformNone struct {
}

// Transform transforms a flow before being stored
func (t *transformNone) Transform(f config.GenericMap) (config.GenericMap, bool) {
	return f, true
}

// NewTransformNone create a new transform
func NewTransformNone() (Transformer, error) {
	logrus.Debugf("entering NewTransformNone")
	return &transformNone{}, nil
}

type Definition struct {
	Type    string
	Generic api.TransformGeneric
	Network api.TransformNetwork
}

type Definitions []Definition
