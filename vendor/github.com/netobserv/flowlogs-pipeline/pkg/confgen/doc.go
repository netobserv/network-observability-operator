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

package confgen

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/extract/aggregate"
)

func (cg *ConfGen) generateVisualizeText(vgs []VisualizationGrafana) string {
	section := ""
	for _, vs := range vgs {
		title := vs.Title
		dashboard := vs.Dashboard
		section = section + fmt.Sprintf("| **Visualized as** | \"%s\" on dashboard `%s` |\n", title, dashboard)
	}

	return section
}

func (cg *ConfGen) generatePromEncodeText(metrics api.PromMetricsItems) string {
	section := ""
	for _, metric := range metrics {
		mType := metric.Type
		name := cg.config.Encode.Prom.Prefix + metric.Name
		section = section + fmt.Sprintf("| **Exposed as** | `%s` of type `%s` |\n", name, mType)
	}

	return section
}

func (cg *ConfGen) generateOperationText(definitions aggregate.Definitions) string {
	section := ""
	for _, definition := range definitions {
		by := strings.Join(definition.By[:], ", ")
		operation := definition.Operation
		recordKey := definition.RecordKey
		if recordKey != "" {
			recordKey = fmt.Sprintf("field `%s`", recordKey)
		}
		section = section + fmt.Sprintf("| **Operation** | aggregate by `%s` and `%s` %s |\n", by, operation, recordKey)
	}

	return section
}

func (cg *ConfGen) generateDoc(fileName string) error {
	doc := ""
	for _, metric := range cg.definitions {
		replacer := strings.NewReplacer("-", " ", "_", " ")
		name := replacer.Replace(filepath.Base(metric.FileName[:len(metric.FileName)-len(filepath.Ext(metric.FileName))]))

		labels := strings.Join(metric.Tags[:], ", ")
		// TODO: add support for multiple operations
		operation := cg.generateOperationText(*metric.AggregateDefinitions)
		expose := cg.generatePromEncodeText(metric.PromEncode.Metrics)
		visualize := cg.generateVisualizeText(metric.Visualization.Grafana)
		doc += fmt.Sprintf(
			`
### %s
| **Description** | %s | 
|:---|:---|
| **Details** | %s | 
| **Usage** | %s | 
| **Tags** | %s |
%s%s%s|||  

`,
			name,
			metric.Description,
			metric.Details,
			metric.Usage,
			labels,
			operation,
			expose,
			visualize,
		)
	}

	header := fmt.Sprintf(`
> Note: this file was automatically generated, to update execute "make generate-configuration"  
> Note: the data was generated from network definitions under the %s folder  
  
# flowlogs-pipeline Metrics  
  
Each table below provides documentation for an exported flowlogs-pipeline metric. 
The documentation describes the metric, the collected information from network flow-logs
and the transformation to generate the exported metric.
  
  

	`, cg.opts.SrcFolder)
	data := fmt.Sprintf("%s\n%s\n", header, doc)
	err := ioutil.WriteFile(fileName, []byte(data), 0664)
	if err != nil {
		return err
	}

	return nil
}
