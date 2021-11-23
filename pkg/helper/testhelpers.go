package helper

import (
	"encoding/json"
)

// AsyncJSON allows converting values to JSON strings at the moment they are printed.
// This type is helpful for printing messages in asynchronous Ginkgo clauses
type AsyncJSON struct {
	Ptr interface{}
}

func (am AsyncJSON) String() string {
	bytes, err := json.Marshal(am.Ptr)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

func Int32Ptr(v int32) *int32 {
	return &v
}
