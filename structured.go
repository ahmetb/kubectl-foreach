package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

type structuredWriter struct {
	format      string
	contextName string
	w           io.Writer
	buf         bytes.Buffer
}

func (s *structuredWriter) Write(p []byte) (int, error) {
	return s.buf.Write(p)
}

func (s *structuredWriter) Close() error {
	data := s.buf.Bytes()
	if len(bytes.TrimSpace(data)) == 0 {
		return nil
	}

	switch s.format {
	case "json":
		return s.closeJSON(data)
	case "yaml":
		return s.closeYAML(data)
	default:
		return fmt.Errorf("unsupported structured format: %s", s.format)
	}
}

func (s *structuredWriter) closeJSON(data []byte) error {
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		_, writeErr := s.w.Write(data)
		return writeErr
	}

	objects := flattenJSON(raw)
	for _, obj := range objects {
		obj["context"] = s.contextName
		line, err := json.Marshal(obj)
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		line = append(line, '\n')
		if _, err := s.w.Write(line); err != nil {
			return fmt.Errorf("failed to write JSON: %w", err)
		}
	}
	return nil
}

func flattenJSON(v any) []map[string]any {
	switch val := v.(type) {
	case map[string]any:
		if items, ok := val["items"]; ok {
			if arr, ok := items.([]any); ok {
				return flattenJSONArray(arr)
			}
		}
		return []map[string]any{val}
	case []any:
		return flattenJSONArray(val)
	default:
		return nil
	}
}

func flattenJSONArray(arr []any) []map[string]any {
	var result []map[string]any
	for _, item := range arr {
		if obj, ok := item.(map[string]any); ok {
			result = append(result, obj)
		}
	}
	return result
}

func (s *structuredWriter) closeYAML(data []byte) error {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	first := true
	for {
		var doc map[string]any
		err := decoder.Decode(&doc)
		if err == io.EOF {
			break
		}
		if err != nil {
			_, writeErr := s.w.Write(data)
			return writeErr
		}
		if doc == nil {
			continue
		}

		doc["context"] = s.contextName

		if !first {
			if _, err := s.w.Write([]byte("---\n")); err != nil {
				return fmt.Errorf("failed to write YAML separator: %w", err)
			}
		}
		first = false

		out, err := yaml.Marshal(doc)
		if err != nil {
			return fmt.Errorf("failed to marshal YAML: %w", err)
		}
		if _, err := s.w.Write(out); err != nil {
			return fmt.Errorf("failed to write YAML: %w", err)
		}
	}
	return nil
}
