package conv_test

import (
	"testing"

	conv "github.com/duh-rpc/openapi-proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertExtractSchemas(t *testing.T) {
	for _, test := range []struct {
		name     string
		given    string
		pkg      string
		expected string
	}{
		{
			name: "extract schemas in YAML insertion order",
			given: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas:
    User:
      type: object
    Product:
      type: object
    Order:
      type: object
`,
			pkg: "testpkg",
			expected: `syntax = "proto3";

package testpkg;

message User {
}

message Product {
}

message Order {
}
`,
		},
		{
			name: "document with no components section",
			given: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
`,
			pkg:      "testpkg",
			expected: "syntax = \"proto3\";\n\npackage testpkg;\n",
		},
		{
			name: "document with empty components/schemas",
			given: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas: {}
`,
			pkg:      "testpkg",
			expected: "syntax = \"proto3\";\n\npackage testpkg;\n",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.given), test.pkg)

			require.NoError(t, err)
			assert.Equal(t, test.expected, string(result))
		})
	}
}

func TestConvertSimpleMessage(t *testing.T) {
	for _, test := range []struct {
		name     string
		given    string
		expected string
		wantErr  string
	}{
		{
			name: "object schema with multiple scalar fields",
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
        userId:
          type: string
        email:
          type: string
        age:
          type: integer
`,
			expected: `syntax = "proto3";

package testpkg;

message User {
  string user_id = 1 [json_name = "userId"];
  string email = 2 [json_name = "email"];
  int32 age = 3 [json_name = "age"];
}
`,
		},
		{
			name: "top-level primitive schema",
			given: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas:
    SimpleString:
      type: string
`,
			wantErr: "only objects and enums supported at top level",
		},
		{
			name: "top-level array schema",
			given: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas:
    StringList:
      type: array
      items:
        type: string
`,
			wantErr: "only objects and enums supported at top level",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.given), "testpkg")

			if test.wantErr != "" {
				require.ErrorContains(t, err, test.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expected, string(result))
		})
	}
}

func TestConvertFieldOrdering(t *testing.T) {
	given := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas:
    Order:
      type: object
      properties:
        orderId:
          type: string
        customerId:
          type: string
        amount:
          type: number
        status:
          type: string
`
	expected := `syntax = "proto3";

package testpkg;

message Order {
  string order_id = 1 [json_name = "orderId"];
  string customer_id = 2 [json_name = "customerId"];
  double amount = 3 [json_name = "amount"];
  string status = 4 [json_name = "status"];
}
`

	result, err := conv.Convert([]byte(given), "testpkg")
	require.NoError(t, err)
	assert.Equal(t, expected, string(result))
}
