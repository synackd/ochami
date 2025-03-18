package format

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"strings"
)

// FormatData marshals arbitrary data into a byte slice formatted as outFormat.
// If a marshalling error occurrs or outFormat is unknown, an error is returned.
//
// Supported values are: json, json-pretty, yaml
func FormatData(data interface{}, outFormat string) ([]byte, error) {
	switch strings.ToLower(outFormat) {
	case "json":
		if bytes, err := json.Marshal(data); err != nil {
			return nil, fmt.Errorf("failed to marshal data into JSON: %w", err)
		} else {
			return bytes, nil
		}
	case "json-pretty":
		if bytes, err := json.MarshalIndent(data, "", "  "); err != nil {
			return nil, fmt.Errorf("failed to marshal data into pretty JSON: %w", err)
		} else {
			return bytes, nil
		}
	case "yaml":
		if bytes, err := yaml.Marshal(data); err != nil {
			return nil, fmt.Errorf("failed to marshal data into YAML: %w", err)
		} else {
			return bytes, nil
		}
	default:
		return nil, fmt.Errorf("unknown data format: %s", outFormat)
	}
}
