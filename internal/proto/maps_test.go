package proto_test

import (
	"testing"

	schema "github.com/duh-rpc/openapi-schema.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConvertAdditionalPropertiesMaps verifies that a `type: object` schema carrying a
// typed `additionalProperties` (and no declared properties) is converted to a proto3
// `map<string, V>` field rather than an empty nested message. This covers the steve
// `Reply.details`/`RunRequest.params` plural-name hard-error case and the slip-stream
// `Event.metadata` singular-name silent-corruption case.
func TestConvertAdditionalPropertiesMaps(t *testing.T) {
	for _, test := range []struct {
		name     string
		given    string
		expected string
		wantErr  string
	}{
		{
			name: "string value map with plural name",
			given: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas:
    Reply:
      type: object
      properties:
        details:
          type: object
          additionalProperties:
            type: string
`,
			expected: `syntax = "proto3";

package testpkg;

option go_package = "github.com/example/proto/v1";

message Reply {
  map<string, string> details = 1 [json_name = "details"];
}

`,
		},
		{
			name: "scalar int64 value map",
			given: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas:
    Counters:
      type: object
      properties:
        counts:
          type: object
          additionalProperties:
            type: integer
            format: int64
`,
			expected: `syntax = "proto3";

package testpkg;

option go_package = "github.com/example/proto/v1";

message Counters {
  map<string, int64> counts = 1 [json_name = "counts"];
}

`,
		},
		{
			name: "ref value map keeps the referenced message reachable",
			given: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas:
    Widget:
      type: object
      properties:
        size:
          type: integer
          format: int32
    Catalog:
      type: object
      properties:
        widgets:
          type: object
          additionalProperties:
            $ref: '#/components/schemas/Widget'
`,
			expected: `syntax = "proto3";

package testpkg;

option go_package = "github.com/example/proto/v1";

message Widget {
  int32 size = 1 [json_name = "size"];
}

message Catalog {
  map<string, Widget> widgets = 1 [json_name = "widgets"];
}

`,
		},
		{
			name: "singular name map is a map field, not an empty nested message",
			given: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas:
    Event:
      type: object
      properties:
        name:
          type: string
        metadata:
          type: object
          additionalProperties:
            type: string
`,
			expected: `syntax = "proto3";

package testpkg;

option go_package = "github.com/example/proto/v1";

message Event {
  string name = 1 [json_name = "name"];
  map<string, string> metadata = 2 [json_name = "metadata"];
}

`,
		},
		{
			name: "object value map is rejected",
			given: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas:
    Bad:
      type: object
      properties:
        nested:
          type: object
          additionalProperties:
            type: object
`,
			wantErr: "map values must be a scalar or $ref",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := schema.Convert([]byte(test.given), schema.ConvertOptions{
				PackageName: "testpkg",
				PackagePath: "github.com/example/proto/v1",
			})

			if test.wantErr != "" {
				require.ErrorContains(t, err, test.wantErr)
				require.Nil(t, result)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, test.expected, string(result.Protobuf))
		})
	}
}

// TestConvertMapDoesNotEmitEmptyMessage locks in the slip-stream silent-corruption fix:
// a singular-named map must not produce an empty `message Metadata {}` alongside the field.
func TestConvertMapDoesNotEmitEmptyMessage(t *testing.T) {
	given := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas:
    Event:
      type: object
      properties:
        metadata:
          type: object
          additionalProperties:
            type: string
`
	result, err := schema.Convert([]byte(given), schema.ConvertOptions{
		PackageName: "testpkg",
		PackagePath: "github.com/example/proto/v1",
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	proto := string(result.Protobuf)
	assert.Contains(t, proto, "map<string, string> metadata = 1")
	assert.NotContains(t, proto, "message Metadata")
}

// TestConvertMapIsDeterministic verifies a map field renders identically across runs,
// keeping its proto field number stable (the fieldmap-lock stability requirement).
func TestConvertMapIsDeterministic(t *testing.T) {
	given := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas:
    Reply:
      type: object
      properties:
        code:
          type: string
        details:
          type: object
          additionalProperties:
            type: string
        params:
          type: object
          additionalProperties:
            type: string
`
	first, err := schema.Convert([]byte(given), schema.ConvertOptions{
		PackageName: "testpkg",
		PackagePath: "github.com/example/proto/v1",
	})
	require.NoError(t, err)
	require.NotNil(t, first)

	second, err := schema.Convert([]byte(given), schema.ConvertOptions{
		PackageName: "testpkg",
		PackagePath: "github.com/example/proto/v1",
	})
	require.NoError(t, err)
	require.NotNil(t, second)

	assert.Equal(t, string(first.Protobuf), string(second.Protobuf))
	assert.Contains(t, string(first.Protobuf), "map<string, string> details = 2")
	assert.Contains(t, string(first.Protobuf), "map<string, string> params = 3")
}
