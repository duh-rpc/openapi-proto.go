package internal_test

import (
	"testing"

	conv "github.com/duh-rpc/openapi-proto.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFieldNumberValidation(t *testing.T) {
	for _, test := range []struct {
		name    string
		given   string
		wantErr string
	}{
		{
			name: "valid field numbers",
			given: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: string
          x-proto-number: 1
        name:
          type: string
          x-proto-number: 2
        email:
          type: string
          x-proto-number: 3
`,
		},
		{
			name: "invalid format - non-numeric string",
			given: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: string
          x-proto-number: abc
`,
			wantErr: "x-proto-number must be a valid integer",
		},
		{
			name: "invalid format - decimal",
			given: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: string
          x-proto-number: 3.14
`,
			wantErr: "x-proto-number must be a valid integer",
		},
		{
			name: "field number 0 rejected",
			given: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: string
          x-proto-number: 0
`,
			wantErr: "x-proto-number must be between 1 and 536870911",
		},
		{
			name: "negative field number rejected",
			given: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: string
          x-proto-number: -1
`,
			wantErr: "x-proto-number must be between 1 and 536870911",
		},
		{
			name: "field number above max rejected",
			given: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: string
          x-proto-number: 536870912
`,
			wantErr: "x-proto-number must be between 1 and 536870911",
		},
		{
			name: "reserved range start rejected",
			given: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: string
          x-proto-number: 19000
`,
			wantErr: "19000 is in reserved range 19000-19999",
		},
		{
			name: "reserved range middle rejected",
			given: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: string
          x-proto-number: 19500
`,
			wantErr: "19500 is in reserved range 19000-19999",
		},
		{
			name: "reserved range end rejected",
			given: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: string
          x-proto-number: 19999
`,
			wantErr: "19999 is in reserved range 19000-19999",
		},
		{
			name: "just below reserved range passes",
			given: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: string
          x-proto-number: 18999
`,
		},
		{
			name: "just above reserved range passes",
			given: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: string
          x-proto-number: 20000
`,
		},
		{
			name: "duplicate field numbers",
			given: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: string
          x-proto-number: 5
        email:
          type: string
          x-proto-number: 5
`,
			wantErr: "duplicate x-proto-number 5",
		},
		{
			name: "maximum valid field number",
			given: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: string
          x-proto-number: 536870911
`,
		},
		{
			name: "schemas without x-proto-number pass",
			given: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: string
        name:
          type: string
`,
		},
		{
			name: "mixed schemas pass in phase 1",
			given: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: string
          x-proto-number: 1
        name:
          type: string
`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := conv.Convert([]byte(test.given), conv.ConvertOptions{
				PackageName: "testpkg",
				PackagePath: "github.com/example/proto/v1",
			})

			if test.wantErr != "" {
				require.ErrorContains(t, err, test.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestExtractFieldNumberValidation(t *testing.T) {
	for _, test := range []struct {
		name     string
		given    string
		expected string
	}{
		{
			name: "valid x-proto-number parsed correctly",
			given: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: string
          x-proto-number: 1
        name:
          type: string
          x-proto-number: 2
`,
			expected: `syntax = "proto3";

package testpkg;

option go_package = "github.com/example/proto/v1";

message User {
  string id = 1 [json_name = "id"];
  string name = 2 [json_name = "name"];
}

`,
		},
		{
			name: "missing x-proto-number uses auto-increment",
			given: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: string
        name:
          type: string
`,
			expected: `syntax = "proto3";

package testpkg;

option go_package = "github.com/example/proto/v1";

message User {
  string id = 1 [json_name = "id"];
  string name = 2 [json_name = "name"];
}

`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.given), conv.ConvertOptions{
				PackageName: "testpkg",
				PackagePath: "github.com/example/proto/v1",
			})
			require.NoError(t, err)
			assert.Equal(t, test.expected, string(result.Protobuf))
		})
	}
}
