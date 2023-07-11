package helper

import "encoding/json"

func FilterDashboardRows(dashboard string, ignoreFlags []string) (string, error) {
	var result map[string]any
	err := json.Unmarshal([]byte(dashboard), &result)
	if err != nil {
		return "", err
	}

	// return dashboard as is if not containing rows
	if result["rows"] == nil {
		return dashboard, nil
	}

	rows := result["rows"].([]any)
	filteredRows := []map[string]any{}
	for _, r := range rows {
		row := r.(map[string]any)

		if row["tags"] != nil {
			t := row["tags"].([]any)
			tags := make([]string, len(t))
			for i := range t {
				tags[i] = t[i].(string)
			}

			// add any row that has tags not included in ignored flags
			if !Intersect(tags, ignoreFlags) {
				filteredRows = append(filteredRows, row)
			}
		} else {
			// add any row that doesn't have tags
			filteredRows = append(filteredRows, row)
		}
	}

	// return empty if dashboard doesn't contains rows anymore
	if len(filteredRows) == 0 {
		return "", nil
	}

	result["rows"] = filteredRows
	bytes, err := json.Marshal(result)
	return string(bytes), err
}
