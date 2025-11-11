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
			name: "mixed schemas rejected with all-or-nothing error",
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
			wantErr: "x-proto-number must be specified on all fields or none",
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

func TestBuildMessageWithFieldNumbers(t *testing.T) {
	for _, test := range []struct {
		name     string
		given    string
		expected string
	}{
		{
			name: "non-sequential field numbers",
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
        email:
          type: string
          x-proto-number: 5
        status:
          type: string
          x-proto-number: 10
`,
			expected: `syntax = "proto3";

package testpkg;

option go_package = "github.com/example/proto/v1";

message User {
  string id = 1 [json_name = "id"];
  string email = 5 [json_name = "email"];
  string status = 10 [json_name = "status"];
}

`,
		},
		{
			name: "property order preserved with custom numbers",
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
        zebra:
          type: string
          x-proto-number: 3
        apple:
          type: string
          x-proto-number: 1
        banana:
          type: string
          x-proto-number: 2
`,
			expected: `syntax = "proto3";

package testpkg;

option go_package = "github.com/example/proto/v1";

message User {
  string zebra = 3 [json_name = "zebra"];
  string apple = 1 [json_name = "apple"];
  string banana = 2 [json_name = "banana"];
}

`,
		},
		{
			name: "large field numbers work",
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
        legacy:
          type: string
          x-proto-number: 100000
`,
			expected: `syntax = "proto3";

package testpkg;

option go_package = "github.com/example/proto/v1";

message User {
  string id = 1 [json_name = "id"];
  string legacy = 100000 [json_name = "legacy"];
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

func TestAllOrNothingValidation(t *testing.T) {
	for _, test := range []struct {
		name    string
		given   string
		wantErr string
	}{
		{
			name: "all fields with x-proto-number passes",
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
			name: "no fields with x-proto-number passes",
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
        email:
          type: string
`,
		},
		{
			name: "1 of 3 fields with x-proto-number fails",
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
        email:
          type: string
`,
			wantErr: "x-proto-number must be specified on all fields or none (found on 1 of 3 fields)",
		},
		{
			name: "2 of 5 fields with x-proto-number fails",
			given: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas:
    Product:
      type: object
      properties:
        id:
          type: string
          x-proto-number: 1
        name:
          type: string
        price:
          type: number
        stock:
          type: integer
          x-proto-number: 5
        status:
          type: string
`,
			wantErr: "x-proto-number must be specified on all fields or none (found on 2 of 5 fields)",
		},
		{
			name: "parent and nested validated independently",
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
        profile:
          type: object
          x-proto-number: 2
          properties:
            name:
              type: string
              x-proto-number: 1
            age:
              type: integer
`,
			wantErr: "x-proto-number must be specified on all fields or none (found on 1 of 2 fields)",
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

func TestBuildNestedMessageWithFieldNumbers(t *testing.T) {
	for _, test := range []struct {
		name     string
		given    string
		expected string
	}{
		{
			name: "nested message with field numbers",
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
        profile:
          type: object
          x-proto-number: 2
          properties:
            name:
              type: string
              x-proto-number: 1
            age:
              type: integer
              x-proto-number: 2
`,
			expected: `syntax = "proto3";

package testpkg;

option go_package = "github.com/example/proto/v1";

message User {
  message Profile {
    string name = 1 [json_name = "name"];
    int32 age = 2 [json_name = "age"];
  }

  string id = 1 [json_name = "id"];
  Profile profile = 2 [json_name = "profile"];
}

`,
		},
		{
			name: "nested messages have independent numbering",
			given: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas:
    Order:
      type: object
      properties:
        id:
          type: string
          x-proto-number: 1
        item:
          type: object
          x-proto-number: 5
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

message Order {
  message Item {
    string id = 1 [json_name = "id"];
    string name = 2 [json_name = "name"];
  }

  string id = 1 [json_name = "id"];
  Item item = 5 [json_name = "item"];
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
