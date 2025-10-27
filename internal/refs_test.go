package internal_test

import (
	"testing"

	conv "github.com/duh-rpc/openapi-proto.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertSchemaReference(t *testing.T) {
	given := `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
components:
  schemas:
    Address:
      type: object
      properties:
        street:
          type: string
        city:
          type: string
    User:
      type: object
      properties:
        name:
          type: string
        address:
          $ref: '#/components/schemas/Address'
`

	expected := `syntax = "proto3";

package testpkg;

option go_package = "github.com/example/proto/v1;testpkg";

message Address {
  string street = 1 [json_name = "street"];
  string city = 2 [json_name = "city"];
}

message User {
  string name = 1 [json_name = "name"];
  Address address = 2 [json_name = "address"];
}

`

	result, err := conv.Convert([]byte(given), conv.ConvertOptions{
		PackageName: "testpkg",
		PackagePath: "github.com/example/proto/v1",
	})
	require.NoError(t, err)
	assert.Equal(t, expected, string(result))
}

func TestConvertMultipleReferences(t *testing.T) {
	given := `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
components:
  schemas:
    Address:
      type: object
      properties:
        street:
          type: string
    User:
      type: object
      properties:
        homeAddress:
          $ref: '#/components/schemas/Address'
        workAddress:
          $ref: '#/components/schemas/Address'
`

	expected := `syntax = "proto3";

package testpkg;

option go_package = "github.com/example/proto/v1;testpkg";

message Address {
  string street = 1 [json_name = "street"];
}

message User {
  Address homeAddress = 1 [json_name = "homeAddress"];
  Address workAddress = 2 [json_name = "workAddress"];
}

`

	result, err := conv.Convert([]byte(given), conv.ConvertOptions{
		PackageName: "testpkg",
		PackagePath: "github.com/example/proto/v1",
	})
	require.NoError(t, err)
	assert.Equal(t, expected, string(result))
}

func TestConvertExternalReference(t *testing.T) {
	given := `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
components:
  schemas:
    User:
      type: object
      properties:
        address:
          $ref: './external.yaml#/components/schemas/Address'
`

	_, err := conv.Convert([]byte(given), conv.ConvertOptions{
		PackageName: "testpkg",
		PackagePath: "github.com/example/proto/v1",
	})
	require.Error(t, err)
	// The error comes from libopenapi build stage indicating the reference cannot be resolved
	assert.Contains(t, err.Error(), "cannot resolve reference")
}
