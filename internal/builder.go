package internal

import (
	"fmt"
	"strings"

	"github.com/duh-rpc/openapi-proto/internal/parser"
	"github.com/pb33f/libopenapi/datamodel/high/base"
)

// Context holds state during conversion
type Context struct {
	Tracker  *NameTracker
	Messages []*ProtoMessage
	Enums    []*ProtoEnum
}

// NewContext creates a new conversion context
func NewContext() *Context {
	return &Context{
		Tracker:  NewNameTracker(),
		Messages: []*ProtoMessage{},
		Enums:    []*ProtoEnum{},
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
			protoType, err := ProtoType(propSchema, propName)
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
