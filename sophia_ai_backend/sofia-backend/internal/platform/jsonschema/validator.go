package jsonschema

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
)

var (
	ErrInvalidSchema   = errors.New("invalid json schema")
	ErrInvalidDocument = errors.New("json does not match schema")
)

type schema struct {
	Type                 string            `json:"type"`
	Required             []string          `json:"required"`
	Properties           map[string]schema `json:"properties"`
	Items                *schema           `json:"items"`
	AdditionalProperties *bool             `json:"additionalProperties"`
	Enum                 []any             `json:"enum"`
}

func Validate(schemaJSON, documentJSON json.RawMessage) error {
	if len(schemaJSON) == 0 || !json.Valid(schemaJSON) {
		return ErrInvalidSchema
	}
	if len(documentJSON) == 0 || !json.Valid(documentJSON) {
		return fmt.Errorf("%w: invalid json", ErrInvalidDocument)
	}

	var s schema
	if err := json.Unmarshal(schemaJSON, &s); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidSchema, err)
	}
	var doc any
	if err := json.Unmarshal(documentJSON, &doc); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidDocument, err)
	}
	if err := validateValue(s, doc, "$"); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidDocument, err)
	}
	return nil
}

func validateValue(s schema, value any, path string) error {
	if s.Type == "" {
		return nil
	}
	if err := validateType(s.Type, value, path); err != nil {
		return err
	}
	if len(s.Enum) > 0 && !containsEnum(s.Enum, value) {
		return fmt.Errorf("%s must be one of enum values", path)
	}
	switch s.Type {
	case "object":
		obj, _ := value.(map[string]any)
		for _, required := range s.Required {
			if _, ok := obj[required]; !ok {
				return fmt.Errorf("%s.%s is required", path, required)
			}
		}
		for key, child := range s.Properties {
			if childValue, ok := obj[key]; ok {
				if err := validateValue(child, childValue, path+"."+key); err != nil {
					return err
				}
			}
		}
		if s.AdditionalProperties != nil && !*s.AdditionalProperties {
			for key := range obj {
				if _, ok := s.Properties[key]; !ok {
					return fmt.Errorf("%s.%s is not allowed", path, key)
				}
			}
		}
	case "array":
		if s.Items == nil {
			return nil
		}
		items, _ := value.([]any)
		for i, item := range items {
			if err := validateValue(*s.Items, item, fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateType(expected string, value any, path string) error {
	switch expected {
	case "object":
		if _, ok := value.(map[string]any); !ok {
			return fmt.Errorf("%s must be object", path)
		}
	case "array":
		if _, ok := value.([]any); !ok {
			return fmt.Errorf("%s must be array", path)
		}
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("%s must be string", path)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("%s must be boolean", path)
		}
	case "number":
		if _, ok := value.(float64); !ok {
			return fmt.Errorf("%s must be number", path)
		}
	case "integer":
		number, ok := value.(float64)
		if !ok || math.Trunc(number) != number {
			return fmt.Errorf("%s must be integer", path)
		}
	default:
		return fmt.Errorf("%s has unsupported schema type %q", path, expected)
	}
	return nil
}

func containsEnum(values []any, candidate any) bool {
	candidateString, candidateIsString := candidate.(string)
	for _, value := range values {
		if value == candidate {
			return true
		}
		if text, ok := value.(string); ok && candidateIsString && strings.EqualFold(text, candidateString) {
			return true
		}
	}
	return false
}
