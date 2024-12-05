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

package decode

import (
	"fmt"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/netobserv/netobserv-ebpf-agent/pkg/decode"
)

type Decoder interface {
	Decode(in []byte) (config.GenericMap, error)
}

func GetDecoder(params api.Decoder) (Decoder, error) {
	switch params.Type {
	case api.DecoderJSON:
		return NewDecodeJSON()
	case api.DecoderProtobuf:
		return decode.NewProtobuf()
	}
	panic(fmt.Sprintf("`decode` type %s not defined", params.Type))
}
