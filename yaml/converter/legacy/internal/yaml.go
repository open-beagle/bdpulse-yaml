package yaml

import (
	"bytes"

	"gopkg.in/yaml.v2"
)

type temporary struct {
	Attributes map[string]interface{} `yaml:",inline"`
	Pipeline   yaml.MapSlice          `yaml:"pipeline"`
}

// this is a helper function that expands merge keys
func expandMergeKeys(b []byte) ([]byte, error) {
	v := new(temporary)
	if err := yaml.Unmarshal(b, v); err != nil {
		return b, err
	}
	o, err := yaml.Marshal(v)
	if err != nil {
		return b, err
	}
	return o, nil
}

func hasMergeKeys(b []byte) bool {
	return bytes.Contains(b, []byte("<<:"))
}
