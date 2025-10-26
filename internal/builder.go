package internal

import (
	"fmt"
	"strings"

	"github.com/duh-rpc/openapi-proto/internal/parser"
	"github.com/pb33f/libopenapi/datamodel/high/base"
)

// Context holds state during conversion
type Context struct {
	Tracker      *NameTracker
	Messages     []*ProtoMessage
	Enums        []*ProtoEnum
	Definitions  []interface{} // Mixed enums and messages in processing order
}

// NewContext creates a new conversion context
func NewContext() *Context {
	return &Context{
		Tracker:     NewNameTracker(),
		Messages:    []*ProtoMessage{},
		Enums:       []*ProtoEnum{},
		Definitions: []interface{}{},
	}
}

// ProtoMessage represents a proto3 message definition
type ProtoMessage struct {
	Name        string
	Description string
	Fields      []*ProtoField
	Nested      []*ProtoMessage
}

// ProtoField represents a proto3 field
type ProtoField struct {
	Name        string
	Type        string
	Number      int
	JSONName    string
	Description string
	Repeated    bool
}

// ProtoEnum represents a proto3 enum definition
type ProtoEnum struct {
	Name        string
	Description string
	Values      []*ProtoEnumValue
}

// ProtoEnumValue represents an enum value
type ProtoEnumValue struct {
	Name   string
	Number int
}

// BuildMessages processes all schemas and returns messages
func BuildMessages(entries []*parser.SchemaEntry, ctx *Context) error {
	for _, entry := range entries {
		// Check if it's an enum schema first
		schema := entry.Proxy.Schema()
		if schema != nil && isEnumSchema(schema) {
			_, err := buildEnum(entry.Name, entry.Proxy, ctx)
			if err != nil {
				return err
			}
			continue
		}

		_, err := buildMessage(entry.Name, entry.Proxy, ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// buildMessage creates a protoMessage from an OpenAPI schema
func buildMessage(name string, proxy *base.SchemaProxy, ctx *Context) (*ProtoMessage, error) {
	schema := proxy.Schema()
	if schema == nil {
		if err := proxy.GetBuildError(); err != nil {
			return nil, fmt.Errorf("schema '%s': failed to resolve schema: %w", name, err)
		}
		return nil, fmt.Errorf("schema '%s': schema is nil", name)
	}

	// Check if it's an object type
	if len(schema.Type) == 0 || !contains(schema.Type, "object") {
		return nil, fmt.Errorf("schema '%s': only objects and enums supported at top level", name)
	}

	msg := &ProtoMessage{
		Name:        ToPascalCase(name),
		Description: schema.Description,
		Fields:      []*ProtoField{},
		Nested:      []*ProtoMessage{},
	}

	// Process properties in YAML order
	if schema.Properties != nil {
		fieldNumber := 1
		for propName, propProxy := range schema.Properties.FromOldest() {
			propSchema := propProxy.Schema()
			if propSchema == nil {
				return nil, fmt.Errorf("schema '%s': property '%s' has nil schema", name, propName)
			}

			protoFieldName := ToSnakeCase(propName)
			protoType, err := ProtoType(propSchema, propName, propProxy, ctx)
			if err != nil {
				return nil, fmt.Errorf("schema '%s': property '%s': %w", name, propName, err)
			}

			field := &ProtoField{
				Name:        protoFieldName,
				Type:        protoType,
				Number:      fieldNumber,
				Description: propSchema.Description,
				Repeated:    false,
			}

			if NeedsJSONName(propName, protoFieldName) {
				field.JSONName = propName
			}

			msg.Fields = append(msg.Fields, field)
			fieldNumber++
		}
	}

	ctx.Messages = append(ctx.Messages, msg)
	ctx.Definitions = append(ctx.Definitions, msg)
	return msg, nil
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}

// isEnumSchema returns true if schema defines an enum
func isEnumSchema(schema *base.Schema) bool {
	return len(schema.Enum) > 0
}

// buildEnum creates a protoEnum from an OpenAPI schema
func buildEnum(name string, proxy *base.SchemaProxy, ctx *Context) (*ProtoEnum, error) {
	schema := proxy.Schema()
	if schema == nil {
		if err := proxy.GetBuildError(); err != nil {
			return nil, fmt.Errorf("schema '%s': failed to resolve schema: %w", name, err)
		}
		return nil, fmt.Errorf("schema '%s': schema is nil", name)
	}

	enumName := ctx.Tracker.UniqueName(ToPascalCase(name))

	enum := &ProtoEnum{
		Name:        enumName,
		Description: schema.Description,
		Values:      []*ProtoEnumValue{},
	}

	// Add UNSPECIFIED value at 0
	unspecifiedName := fmt.Sprintf("%s_UNSPECIFIED", strings.ToUpper(ToSnakeCase(enumName)))
	enum.Values = append(enum.Values, &ProtoEnumValue{
		Name:   unspecifiedName,
		Number: 0,
	})

	// Add original enum values starting at 1
	for i, value := range schema.Enum {
		// Extract the actual value from yaml.Node
		// The Value field contains the string representation
		var strValue string
		if value != nil {
			strValue = value.Value
		}
		valueName := ToEnumValueName(enumName, strValue)
		enum.Values = append(enum.Values, &ProtoEnumValue{
			Name:   valueName,
			Number: i + 1,
		})
	}

	ctx.Enums = append(ctx.Enums, enum)
	ctx.Definitions = append(ctx.Definitions, enum)
	return enum, nil
}
