package api

import (
	"encoding/json"
	"fmt"
	"time"
)

// Duration is a wrapper of time.Duration that allows json marshaling.
// https://stackoverflow.com/a/48051946/2749989
type Duration struct {
	time.Duration
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)
		return nil
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("invalid duration %v", value)
	}
}

func (d Duration) MarshalYAML() (interface{}, error) {
	return d.String(), nil
}

func (d *Duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var durationStr string
	err := unmarshal(&durationStr)
	if err != nil {
		return err
	}
	d.Duration, err = time.ParseDuration(durationStr)
	if err != nil {
		return err
	}
	return nil
}
