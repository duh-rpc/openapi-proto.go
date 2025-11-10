package conv

import (
	"fmt"

	"github.com/duh-rpc/openapi-proto.go/internal"
	"github.com/duh-rpc/openapi-proto.go/internal/parser"
)

// ConvertResult contains the outputs from converting OpenAPI to proto3 and Go code
type ConvertResult struct {
	Protobuf []byte
	Golang   []byte
	TypeMap  map[string]*TypeInfo
}

// TypeInfo contains metadata about where a type is generated and why
type TypeInfo struct {
	Location TypeLocation
	Reason   string
}

// TypeLocation indicates whether a type is generated as proto or golang
type TypeLocation string

const (
	TypeLocationProto  TypeLocation = "proto"
	TypeLocationGolang TypeLocation = "golang"
)

// ConvertOptions configures the conversion from OpenAPI to Protocol Buffers
type ConvertOptions struct {
	// PackageName is the name of the generated proto3 package (e.g. "api")
	PackageName string
	// PackagePath is the path of the generated proto3 package (e.g. "github.com/myorg/proto/v1/api")
	PackagePath string
	// GoPackagePath is the path for generated Go code (defaults to PackagePath if empty)
	GoPackagePath string
}

// Convert converts OpenAPI 3.0 schemas to Protocol Buffer 3 format.
// It takes OpenAPI specification bytes (YAML or JSON) and conversion options,
// and returns a ConvertResult containing proto3 output, Go output, and type metadata.
//
// Field names are preserved from the OpenAPI schema when they meet proto3 syntax
// requirements. Invalid characters (hyphens, dots, spaces) are replaced with
// underscores. All fields include json_name annotations for correct JSON mapping.
//
// Examples:
//   - HTTPStatus → HTTPStatus [json_name = "HTTPStatus"]
//   - userId → userId [json_name = "userId"]
//   - status-code → status_code [json_name = "status-code"]
//
// The function validates inputs, parses the OpenAPI document, extracts schemas,
// and generates corresponding proto3 message definitions.
//
// Returns an error if:
//   - openapi is empty
//   - opts.PackageName is empty
//   - opts.PackagePath is empty
//   - the OpenAPI document is invalid or not version 3.x
//   - any schema contains unsupported features
func Convert(openapi []byte, opts ConvertOptions) (*ConvertResult, error) {
	if len(openapi) == 0 {
		return nil, fmt.Errorf("openapi input cannot be empty")
	}

	if opts.PackageName == "" {
		return nil, fmt.Errorf("package name cannot be empty")
	}

	if opts.PackagePath == "" {
		return nil, fmt.Errorf("package path cannot be empty")
	}

	// Default GoPackagePath to PackagePath if not provided
	if opts.GoPackagePath == "" {
		opts.GoPackagePath = opts.PackagePath
	}

	doc, err := parser.ParseDocument(openapi)
	if err != nil {
		return nil, err
	}

	schemas, err := doc.Schemas()
	if err != nil {
		return nil, err
	}

	ctx := internal.NewContext()
	graph, err := internal.BuildMessages(schemas, ctx)
	if err != nil {
		return nil, err
	}

	// Compute transitive closure to classify types
	goTypes, protoTypes, reasons := graph.ComputeTransitiveClosure()

	// Build TypeMap using classification results
	typeMap := buildTypeMap(goTypes, protoTypes, reasons)

	// Generate proto for proto-only types
	// Skip proto generation only if there are Go types but no proto types
	var protoBytes []byte
	if len(protoTypes) > 0 || len(goTypes) == 0 {
		protoMessages := filterProtoMessages(ctx.Messages, protoTypes)
		// Create new context with filtered messages
		protoCtx := internal.NewContext()
		protoCtx.Messages = protoMessages
		protoCtx.Enums = ctx.Enums
		protoCtx.Definitions = filterProtoDefinitions(ctx.Definitions, protoTypes)
		protoCtx.UsesTimestamp = ctx.UsesTimestamp

		protoBytes, err = internal.Generate(opts.PackageName, opts.PackagePath, protoCtx)
		if err != nil {
			return nil, err
		}
	}

	// Generate Go for Go-only types
	var goBytes []byte
	if len(goTypes) > 0 {
		goCtx := internal.NewGoContext(internal.ExtractPackageName(opts.GoPackagePath))
		err := internal.BuildGoStructs(schemas, goTypes, graph, goCtx)
		if err != nil {
			return nil, err
		}
		goBytes, err = internal.GenerateGo(goCtx)
		if err != nil {
			return nil, err
		}
	}

	return &ConvertResult{
		Protobuf: protoBytes,
		Golang:   goBytes,
		TypeMap:  typeMap,
	}, nil
}

// buildTypeMap creates a TypeMap from dependency graph classification results
func buildTypeMap(goTypes, protoTypes map[string]bool, reasons map[string]string) map[string]*TypeInfo {
	typeMap := make(map[string]*TypeInfo)

	// Add Go types
	for name := range goTypes {
		typeMap[name] = &TypeInfo{
			Location: TypeLocationGolang,
			Reason:   reasons[name],
		}
	}

	// Add Proto types
	for name := range protoTypes {
		typeMap[name] = &TypeInfo{
			Location: TypeLocationProto,
			Reason:   "",
		}
	}

	return typeMap
}

// filterProtoMessages removes messages marked as Go-only from proto output
func filterProtoMessages(messages []*internal.ProtoMessage, protoTypes map[string]bool) []*internal.ProtoMessage {
	filtered := make([]*internal.ProtoMessage, 0, len(protoTypes))

	for _, msg := range messages {
		// Only include messages that are in protoTypes set (using original schema name)
		if protoTypes[msg.OriginalSchema] {
			filtered = append(filtered, msg)
		}
	}

	return filtered
}

// filterProtoDefinitions removes definitions marked as Go-only from proto output
func filterProtoDefinitions(definitions []interface{}, protoTypes map[string]bool) []interface{} {
	filtered := make([]interface{}, 0)

	for _, def := range definitions {
		// Check if it's a ProtoMessage and filter accordingly
		if msg, ok := def.(*internal.ProtoMessage); ok {
			if protoTypes[msg.OriginalSchema] {
				filtered = append(filtered, def)
			}
		} else {
			// Keep enums and other definitions
			filtered = append(filtered, def)
		}
	}

	return filtered
}
