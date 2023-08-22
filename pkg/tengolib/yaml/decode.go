package yaml

import (
	"encoding/json"

	"github.com/d5/tengo/v2"
	tjson "github.com/d5/tengo/v2/stdlib/json"
	"gopkg.in/yaml.v3"
)

// Decode parses the YAML-encoded data and returns the result object.
func Decode(data []byte) (tengo.Object, error) {
	var m map[string]any
	err := yaml.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}

	// Just go the easy way by encoding in JSON and using Tengo's JSON decoder.
	// It's not as efficient as converting the map[string]any to a *tengo.Map.
	// Parsing the YAML directly would be painful because the YAML spec is a headache.
	data, err = json.Marshal(m)
	if err != nil {
		return nil, err
	}

	return tjson.Decode(data)
}
