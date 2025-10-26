package internal

import (
	"fmt"
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/base"
)

// ProtoType returns the proto3 type for an OpenAPI schema.
// Returns type name, whether it's repeated, and error.
// For inline enums and objects, hoists them appropriately in the context.
// parentMsg is used for nested messages (can be nil for top-level).
func ProtoType(schema *base.Schema, propertyName string, propProxy *base.SchemaProxy, ctx *Context, parentMsg *ProtoMessage) (string, bool, error) {
	// Check if it's a reference first
	if propProxy.IsReference() {
		ref := propProxy.GetReference()

		// Try to resolve the reference (libopenapi handles internal refs automatically)
		resolvedSchema := propProxy.Schema()
		if resolvedSchema == nil {
			// Check if there's a build error (e.g., external reference)
			if err := propProxy.GetBuildError(); err != nil {
				return "", false, fmt.Errorf("property '%s' references external file or unresolvable reference: %w", propertyName, err)
			}
			return "", false, fmt.Errorf("property '%s' has unresolved reference", propertyName)
		}

		// Extract the schema name from the reference
		typeName, err := extractReferenceName(ref)
		if err != nil {
			return "", false, fmt.Errorf("property '%s': %w", propertyName, err)
		}
		return typeName, false, nil
	}

	// Check if it's an array first
	if len(schema.Type) > 0 && contains(schema.Type, "array") {
		itemType, err := ResolveArrayItemType(schema, propertyName, propProxy, ctx, parentMsg)
		if err != nil {
			return "", false, err
		}
		return itemType, true, nil
	}

	// Check if it's an inline object
	if len(schema.Type) > 0 && contains(schema.Type, "object") {
		// Build nested message
		nestedMsg, err := buildNestedMessage(propertyName, propProxy, ctx, parentMsg)
		if err != nil {
			return "", false, err
		}
		return nestedMsg.Name, false, nil
	}

	// Check if it's an enum
	if isEnumSchema(schema) {
		// Hoist inline enum to top-level
		enumName := ToPascalCase(propertyName)
		_, err := buildEnum(enumName, propProxy, ctx)
		if err != nil {
			return "", false, err
		}
		return enumName, false, nil
	}

	if len(schema.Type) == 0 {
		return "", false, fmt.Errorf("property must have type or $ref")
	}

	if len(schema.Type) > 1 {
		return "", false, fmt.Errorf("multi-type properties not supported")
	}

	typ := schema.Type[0]
	format := schema.Format

	scalarType, err := MapScalarType(typ, format)
	return scalarType, false, err
}

// MapScalarType maps OpenAPI type+format to proto3 scalar type.
func MapScalarType(typ, format string) (string, error) {
	switch typ {
	case "integer":
		if format == "int64" {
			return "int64", nil
		}
		return "int32", nil

	case "number":
		if format == "float" {
			return "float", nil
		}
		return "double", nil

	case "string":
		if format == "byte" || format == "binary" {
			return "bytes", nil
		}
		return "string", nil

	case "boolean":
		return "bool", nil

	default:
		return "", fmt.Errorf("unsupported type: %s", typ)
	}
}

// ResolveArrayItemType determines the proto3 type for array items.
// Returns type name, whether it's repeated, and error.
// For inline objects/enums: validates property name is not plural.
func ResolveArrayItemType(schema *base.Schema, propertyName string, propProxy *base.SchemaProxy, ctx *Context, parentMsg *ProtoMessage) (string, error) {
	// Check if Items is defined
	if schema.Items == nil || schema.Items.A == nil {
		return "", fmt.Errorf("array must have items defined")
	}

	itemsProxy := schema.Items.A
	itemsSchema := itemsProxy.Schema()
	if itemsSchema == nil {
		if err := itemsProxy.GetBuildError(); err != nil {
			return "", fmt.Errorf("failed to resolve array items: %w", err)
		}
		return "", fmt.Errorf("array items schema is nil")
	}

	// Check for nested arrays
	if len(itemsSchema.Type) > 0 && contains(itemsSchema.Type, "array") {
		return "", fmt.Errorf("nested arrays not supported")
	}

	// Check if it's a reference
	if itemsProxy.IsReference() {
		ref := itemsProxy.GetReference()
		if ref != "" {
			// Extract the last segment of the reference path
			parts := strings.Split(ref, "/")
			if len(parts) > 0 {
				return parts[len(parts)-1], nil
			}
		}
		return "", fmt.Errorf("invalid reference format")
	}

	// Check if it's an inline enum
	if isEnumSchema(itemsSchema) {
		// Validate property name is not plural
		if strings.HasSuffix(propertyName, "es") {
			return "", fmt.Errorf("cannot derive enum name from plural array property '%s'; use singular form or $ref", propertyName)
		}
		if strings.HasSuffix(propertyName, "s") {
			return "", fmt.Errorf("cannot derive enum name from plural array property '%s'; use singular form or $ref", propertyName)
		}

		// Hoist inline enum to top-level
		enumName := ToPascalCase(propertyName)
		_, err := buildEnum(enumName, itemsProxy, ctx)
		if err != nil {
			return "", err
		}
		return enumName, nil
	}

	// Check if it's an inline object
	if len(itemsSchema.Type) > 0 && contains(itemsSchema.Type, "object") {
		// Validate property name is not plural
		if strings.HasSuffix(propertyName, "es") {
			return "", fmt.Errorf("cannot derive message name from plural array property '%s'; use singular form or $ref", propertyName)
		}
		if strings.HasSuffix(propertyName, "s") {
			return "", fmt.Errorf("cannot derive message name from plural array property '%s'; use singular form or $ref", propertyName)
		}

		// Build nested message for inline object in array
		nestedMsg, err := buildNestedMessage(propertyName, itemsProxy, ctx, parentMsg)
		if err != nil {
			return "", err
		}
		return nestedMsg.Name, nil
	}

	// It's a scalar type
	if len(itemsSchema.Type) == 0 {
		return "", fmt.Errorf("array items must have a type")
	}

	itemType := itemsSchema.Type[0]
	format := itemsSchema.Format
	return MapScalarType(itemType, format)
}

// extractReferenceName extracts the schema name from a reference string.
// Example: "#/components/schemas/Address" → "Address"
func extractReferenceName(ref string) (string, error) {
	if ref == "" {
		return "", fmt.Errorf("reference string is empty")
	}

	// Split by '/' and take the last segment
	parts := strings.Split(ref, "/")
	if len(parts) == 0 {
		return "", fmt.Errorf("invalid reference format: %s", ref)
	}

	name := parts[len(parts)-1]
	if name == "" {
		return "", fmt.Errorf("reference has empty name segment: %s", ref)
	}

	return name, nil
}
