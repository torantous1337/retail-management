package services

import (
	"errors"
	"testing"

	"github.com/torantous1337/retail-management/internal/core/domain"
)

func TestValidateProperties_NilCategory(t *testing.T) {
	err := ValidateProperties(nil, map[string]interface{}{"key": "val"})
	if err != nil {
		t.Fatalf("expected nil error for nil category, got %v", err)
	}
}

func TestValidateProperties_RequiredFieldMissing(t *testing.T) {
	cat := &domain.Category{
		AttributeDefinitions: []domain.AttributeDefinition{
			{Key: "voltage", Type: "number", Required: true, Unit: "Volts"},
		},
	}
	err := ValidateProperties(cat, map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error for missing required field")
	}
	if !errors.Is(err, ErrInvalidProperty) {
		t.Fatalf("expected ErrInvalidProperty, got %v", err)
	}
}

func TestValidateProperties_RequiredFieldPresent(t *testing.T) {
	cat := &domain.Category{
		AttributeDefinitions: []domain.AttributeDefinition{
			{Key: "voltage", Type: "number", Required: true, Unit: "Volts"},
		},
	}
	err := ValidateProperties(cat, map[string]interface{}{"voltage": 220.0})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidateProperties_StringType(t *testing.T) {
	cat := &domain.Category{
		AttributeDefinitions: []domain.AttributeDefinition{
			{Key: "brand", Type: "string", Required: true},
		},
	}

	// Valid string
	err := ValidateProperties(cat, map[string]interface{}{"brand": "Acme"})
	if err != nil {
		t.Fatalf("expected no error for valid string, got %v", err)
	}

	// Invalid: number instead of string
	err = ValidateProperties(cat, map[string]interface{}{"brand": 123.0})
	if !errors.Is(err, ErrInvalidProperty) {
		t.Fatalf("expected ErrInvalidProperty for non-string, got %v", err)
	}
}

func TestValidateProperties_NumberType(t *testing.T) {
	cat := &domain.Category{
		AttributeDefinitions: []domain.AttributeDefinition{
			{Key: "wattage", Type: "number", Required: true, Unit: "Watts"},
		},
	}

	// Valid float64 (JSON default for numbers)
	err := ValidateProperties(cat, map[string]interface{}{"wattage": 100.0})
	if err != nil {
		t.Fatalf("expected no error for valid number, got %v", err)
	}

	// Valid numeric string
	err = ValidateProperties(cat, map[string]interface{}{"wattage": "100"})
	if err != nil {
		t.Fatalf("expected no error for numeric string, got %v", err)
	}

	// Invalid: non-numeric string
	err = ValidateProperties(cat, map[string]interface{}{"wattage": "abc"})
	if !errors.Is(err, ErrInvalidProperty) {
		t.Fatalf("expected ErrInvalidProperty for non-numeric, got %v", err)
	}
}

func TestValidateProperties_BooleanType(t *testing.T) {
	cat := &domain.Category{
		AttributeDefinitions: []domain.AttributeDefinition{
			{Key: "is_organic", Type: "boolean", Required: true},
		},
	}

	// Valid boolean
	err := ValidateProperties(cat, map[string]interface{}{"is_organic": true})
	if err != nil {
		t.Fatalf("expected no error for valid boolean, got %v", err)
	}

	// Invalid: string instead of bool
	err = ValidateProperties(cat, map[string]interface{}{"is_organic": "yes"})
	if !errors.Is(err, ErrInvalidProperty) {
		t.Fatalf("expected ErrInvalidProperty for non-boolean, got %v", err)
	}
}

func TestValidateProperties_SelectType(t *testing.T) {
	cat := &domain.Category{
		AttributeDefinitions: []domain.AttributeDefinition{
			{Key: "color", Type: "select", Required: true, Options: []string{"Red", "Blue", "Green"}},
		},
	}

	// Valid option
	err := ValidateProperties(cat, map[string]interface{}{"color": "Red"})
	if err != nil {
		t.Fatalf("expected no error for valid select option, got %v", err)
	}

	// Invalid option
	err = ValidateProperties(cat, map[string]interface{}{"color": "Yellow"})
	if !errors.Is(err, ErrInvalidProperty) {
		t.Fatalf("expected ErrInvalidProperty for invalid select option, got %v", err)
	}

	// Invalid type for select (not a string)
	err = ValidateProperties(cat, map[string]interface{}{"color": 123})
	if !errors.Is(err, ErrInvalidProperty) {
		t.Fatalf("expected ErrInvalidProperty for non-string select, got %v", err)
	}
}

func TestValidateProperties_OptionalFieldMissing(t *testing.T) {
	cat := &domain.Category{
		AttributeDefinitions: []domain.AttributeDefinition{
			{Key: "notes", Type: "string", Required: false},
		},
	}

	// Optional field not provided - should be fine
	err := ValidateProperties(cat, map[string]interface{}{})
	if err != nil {
		t.Fatalf("expected no error for missing optional field, got %v", err)
	}
}

func TestValidateProperties_MultipleAttributes(t *testing.T) {
	cat := &domain.Category{
		Name: "Electrical",
		AttributeDefinitions: []domain.AttributeDefinition{
			{Key: "voltage", Type: "number", Required: true, Unit: "Volts"},
			{Key: "color", Type: "select", Required: true, Options: []string{"Red", "Blue"}},
			{Key: "brand", Type: "string", Required: false},
		},
	}

	// All valid
	props := map[string]interface{}{
		"voltage": 220.0,
		"color":   "Red",
		"brand":   "Acme",
	}
	err := ValidateProperties(cat, props)
	if err != nil {
		t.Fatalf("expected no error for all valid properties, got %v", err)
	}

	// Missing required "color"
	props2 := map[string]interface{}{
		"voltage": 220.0,
	}
	err = ValidateProperties(cat, props2)
	if !errors.Is(err, ErrInvalidProperty) {
		t.Fatalf("expected ErrInvalidProperty for missing required field, got %v", err)
	}
}
