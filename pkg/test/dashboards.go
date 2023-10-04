package test

import (
	"encoding/json"
)

type Dashboard struct {
	Rows []struct {
		Panels []struct {
			Targets []struct {
				Expr         string `json:"expr"`
				LegendFormat string `json:"legendFormat"`
			} `json:"targets"`
			Title string `json:"title"`
		} `json:"panels"`
		Title string `json:"title"`
	} `json:"rows"`
	Title string `json:"title"`
}

func DashboardFromBytes(b []byte) (*Dashboard, error) {
	var d Dashboard
	err := json.Unmarshal(b, &d)
	return &d, err
}

func (d *Dashboard) Titles() []string {
	var titles []string
	for _, r := range d.Rows {
		titles = append(titles, r.Title)
	}
	return titles
}
