package domain

// AttributeDefinition defines a single rule for a product property.
type AttributeDefinition struct {
	Key      string   // e.g., "voltage"
	Type     string   // "string", "number", "boolean", "select"
	Required bool
	Options  []string // for "select" types, e.g., ["Red", "Blue"]
	Unit     string   // e.g., "Volts", "Ohms"
}

// Category defines the blueprint (schema) for product properties.
type Category struct {
	ID                   string
	Name                 string // e.g., "Electrical", "Liquor"
	AttributeDefinitions []AttributeDefinition
}
