package conv

import (
	"fmt"

	"github.com/duh-rpc/openapi-proto/internal"
	"github.com/duh-rpc/openapi-proto/internal/parser"
)

// Convert converts OpenAPI 3.0 schemas to Protocol Buffer 3 format.
// It takes OpenAPI specification bytes (YAML or JSON) and a protobuf package name,
// and returns the generated proto3 file content as bytes.
//
// The function validates inputs, parses the OpenAPI document, extracts schemas,
// and generates corresponding proto3 message definitions.
//
// Returns an error if:
//   - openapi is empty
//   - packageName is empty
//   - the OpenAPI document is invalid or not version 3.x
//   - any schema contains unsupported features
func Convert(openapi []byte, packageName string) ([]byte, error) {
	if len(openapi) == 0 {
		return nil, fmt.Errorf("openapi input cannot be empty")
	}

	if packageName == "" {
		return nil, fmt.Errorf("package name cannot be empty")
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

	return internal.Generate(packageName, ctx.Messages, ctx.Enums)
}
