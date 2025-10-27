package conv

import (
	"fmt"

	"github.com/duh-rpc/openapi-proto.go/internal"
	"github.com/duh-rpc/openapi-proto.go/internal/parser"
)

// ConvertOptions configures the conversion from OpenAPI to Protocol Buffers
type ConvertOptions struct {
	// PackageName is the name of the generated proto3 package (e.g. "api")
	PackageName string
	// PackagePath is the path of the generated proto3 package (e.g. "github.com/myorg/proto/v1/api")
	PackagePath string
}

// Convert converts OpenAPI 3.0 schemas to Protocol Buffer 3 format.
// It takes OpenAPI specification bytes (YAML or JSON) and conversion options,
// and returns the generated proto3 file content as bytes.
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
func Convert(openapi []byte, opts ConvertOptions) ([]byte, error) {
	if len(openapi) == 0 {
		return nil, fmt.Errorf("openapi input cannot be empty")
	}

	if opts.PackageName == "" {
		return nil, fmt.Errorf("package name cannot be empty")
	}

	if opts.PackagePath == "" {
		return nil, fmt.Errorf("package path cannot be empty")
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
	err = internal.BuildMessages(schemas, ctx)
	if err != nil {
		return nil, err
	}

	return internal.Generate(opts.PackageName, opts.PackagePath, ctx.Messages, ctx.Enums, ctx.Definitions, ctx.UsesTimestamp)
}
