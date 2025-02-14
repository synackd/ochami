package utils

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"strings"
)

// FormatOutput formats the output as JSON or YAML
func FormatOutput(output interface{}, format string) ([]byte, error) {
	switch strings.ToLower(format) {
	case "json":
		if bytes, err := json.Marshal(output); err != nil {
			return nil, fmt.Errorf("failed to marshal output into JSON: %w", err)
		} else {
			return bytes, nil
		}
	case "json-pretty":
		if bytes, err := json.MarshalIndent(output, "", "  "); err != nil {
			return nil, fmt.Errorf("failed to marshal output into JSON: %w", err)
		} else {
			return bytes, nil
		}
	case "yaml":
		if bytes, err := yaml.Marshal(output); err != nil {
			return nil, fmt.Errorf("failed to marshal output into YAML: %w", err)
		} else {
			return bytes, nil
		}
	default:
		return nil, fmt.Errorf("unknown output format: %s", format)
	}
}
