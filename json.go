package main

import (
    "encoding/json"
    "fmt"
    "strings"
)

// JSONProcessor handles JSON manipulation operations
type JSONProcessor struct {
    // any is an alias of: interface{}
    data any
    lines []string
}

// NewJSONProcessor creates a new processor from JSON bytes
func NewJSONProcessor(jsonData []byte) (*JSONProcessor, error) {
	var data any 
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	processor := &JSONProcessor{data: data}
	if err := processor.updateLines(); err != nil {
		return nil, err
	}

	return processor, nil
}

// updateLines converts the current JSON data to pretty-printed string array
func (json_processor *JSONProcessor) updateLines() error {
	prettyJSON, err := json.MarshalIndent(json_processor.data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	json_processor.lines = strings.Split(string(prettyJSON), "\n")
	return nil
}

// QueryField extracts a field value using dot notation (e.g., "merchant.category_code")
func (jp *JSONProcessor) QueryField(path string) (any, error) {
	parts := strings.Split(path, ".")
	current := jp.data

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]any:
			var ok bool
			current, ok = v[part]
			if !ok {
				return nil, fmt.Errorf("field '%s' not found", part)
			}
		default:
			return nil, fmt.Errorf("cannot access field '%s' on non-object type", part)
		}
	}

	return current, nil
}

