package dashboards

import (
	_ "embed"
	"encoding/json"
	"strings"
)

//go:embed infra_health_dashboard.json
var healthDashboardEmbed string

func CreateHealthDashboard(netobsNs string, metrics []string) (string, error) {
	var result map[string]any
	err := json.Unmarshal([]byte(healthDashboardEmbed), &result)
	if err != nil {
		return "", err
	}

	// return dashboard as is if not containing rows
	if result["rows"] == nil {
		return healthDashboardEmbed, nil
	}

	rows := result["rows"].([]any)
	filteredRows := []map[string]any{}
	for _, r := range rows {
		row := r.(map[string]any)

		if isRowPresent(row, metrics) {
			filteredRows = append(filteredRows, row)
		}
	}

	// return empty if dashboard doesn't contains rows anymore
	if len(filteredRows) == 0 {
		return "", nil
	}

	result["rows"] = filteredRows
	bytes, err := json.Marshal(result)
	return strings.ReplaceAll(string(bytes), "$NETOBSERV_NS", netobsNs), err
}

func hasTag(item map[string]any, search string) bool {
	if item["tags"] == nil {
		return false
	}
	tags := item["tags"].([]any)
	for _, t := range tags {
		tag := t.(string)
		if tag == search {
			return true
		}
	}
	return false
}

func isRowPresent(row map[string]any, metrics []string) bool {
	if !hasTag(row, "dynamic") {
		return true
	}
	panels := row["panels"].([]any)
	for _, p := range panels {
		panel := p.(map[string]any)
		targets := panel["targets"].([]any)
		for _, t := range targets {
			target := t.(map[string]any)
			expr := target["expr"].(string)
			for _, metric := range metrics {
				if strings.Contains(expr, metric) {
					return true
				}
			}
		}
	}
	return false
}
