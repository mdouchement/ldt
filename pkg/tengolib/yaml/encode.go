package yaml

import (
	"encoding/json"

	"github.com/d5/tengo/v2"
	tjson "github.com/d5/tengo/v2/stdlib/json"
	"gopkg.in/yaml.v3"
)

// Encode returns the JSON encoding of the object.
func Encode(o tengo.Object) ([]byte, error) {
	// Just go the easy way by encoding in JSON and using Tengo's JSON encoder.
	data, err := tjson.Encode(o)
	if err != nil {
		return nil, err
	}

	var m map[string]any
	err = json.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}

	return yaml.Marshal(m)
}
