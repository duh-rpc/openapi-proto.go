package internal

import "fmt"

// SchemaError creates an error with schema context.
// Format: schema '<name>': <message>
func SchemaError(schemaName, message string) error {
	return fmt.Errorf("schema '%s': %s", schemaName, message)
}

// PropertyError creates an error with schema and property context.
// Format: schema '<schema>': property '<prop>' <message>
func PropertyError(schemaName, propertyName, message string) error {
	return fmt.Errorf("schema '%s': property '%s' %s", schemaName, propertyName, message)
}

// UnsupportedError creates an error for unsupported features.
// Format: schema '<schema>': property '<prop>' uses '<feature>' which is not supported
func UnsupportedError(schemaName, propertyName, feature string) error {
	return fmt.Errorf("schema '%s': property '%s' uses '%s' which is not supported", schemaName, propertyName, feature)
}

// UnsupportedSchemaError creates an error for unsupported features at the schema level.
// Format: schema '<name>': uses '<feature>' which is not supported
func UnsupportedSchemaError(schemaName, feature string) error {
	return fmt.Errorf("schema '%s': uses '%s' which is not supported", schemaName, feature)
}
