package internal

import (
	"fmt"
	"github.com/pb33f/libopenapi/datamodel/high/base"
)

// ProtoType returns the proto3 type for an OpenAPI schema.
// Returns type name and error.
// For inline enums, hoists them to top-level in the context.
func ProtoType(schema *base.Schema, propertyName string, propProxy *base.SchemaProxy, ctx *Context) (string, error) {
	// Check if it's an enum first
	if isEnumSchema(schema) {
		// Hoist inline enum to top-level
		enumName := ToPascalCase(propertyName)
		_, err := buildEnum(enumName, propProxy, ctx)
		if err != nil {
			return "", err
		}
		return enumName, nil
	}

	if len(schema.Type) == 0 {
		return "", fmt.Errorf("property must have type or $ref")
	}

	if len(schema.Type) > 1 {
		return "", fmt.Errorf("multi-type properties not supported")
	}

	typ := schema.Type[0]
	format := schema.Format

	return MapScalarType(typ, format)
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
