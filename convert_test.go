package conv_test

import (
	"testing"

	conv "github.com/duh-rpc/openapi-proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertBasics(t *testing.T) {
	for _, test := range []struct {
		name     string
		given    []byte
		pkg      string
		expected string
		wantErr  string
	}{
		{
			name:    "empty openapi bytes",
			given:   []byte{},
			pkg:     "testpkg",
			wantErr: "openapi input cannot be empty",
		},
		{
			name:    "empty package name",
			given:   []byte("openapi: 3.0.0"),
			pkg:     "",
			wantErr: "package name cannot be empty",
		},
		{
			name:    "invalid YAML syntax",
			given:   []byte("this is not valid: [yaml"),
			pkg:     "testpkg",
			wantErr: "failed to parse OpenAPI document",
		},
		{
			name: "valid minimal OpenAPI 3.0",
			given: []byte(`openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
`),
			pkg:      "testpkg",
			expected: "syntax = \"proto3\";\n\npackage testpkg;\n",
		},
		{
			name: "OpenAPI 2.0 Swagger",
			given: []byte(`swagger: "2.0"
info:
  title: Test API
  version: 1.0.0
paths: {}
`),
			pkg:     "testpkg",
			wantErr: "supplied spec is a different version",
		},
		{
			name: "valid JSON OpenAPI",
			given: []byte(`{
  "openapi": "3.0.0",
  "info": {
    "title": "Test API",
    "version": "1.0.0"
  },
  "paths": {}
}`),
			pkg:      "testpkg",
			expected: "syntax = \"proto3\";\n\npackage testpkg;\n",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert(test.given, test.pkg)

			if test.wantErr != "" {
				require.ErrorContains(t, err, test.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expected, string(result))
		})
	}
}

func TestConvertParseDocument(t *testing.T) {
	for _, test := range []struct {
		name     string
		given    string
		pkg      string
		expected string
		wantErr  string
	}{
		{
			name: "parse valid OpenAPI 3.0 YAML",
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
			name: "parse valid OpenAPI 3.0 JSON",
			given: `{
  "openapi": "3.0.0",
  "info": {
    "title": "Test API",
    "version": "1.0.0"
  },
  "paths": {}
}`,
			pkg:      "testpkg",
			expected: "syntax = \"proto3\";\n\npackage testpkg;\n",
		},
		{
			name:    "non-OpenAPI document",
			given:   `title: Some Random YAML`,
			pkg:     "testpkg",
			wantErr: "spec type not supported",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.given), test.pkg)

			if test.wantErr != "" {
				require.ErrorContains(t, err, test.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expected, string(result))
		})
	}
}

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

func TestConvertScalarTypes(t *testing.T) {
	for _, test := range []struct {
		name     string
		given    string
		expected string
		wantErr  string
	}{
		{
			name: "all scalar type mappings",
			given: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas:
    AllTypes:
      type: object
      properties:
        int32Field:
          type: integer
          format: int32
        int64Field:
          type: integer
          format: int64
        floatField:
          type: number
          format: float
        doubleField:
          type: number
          format: double
        stringField:
          type: string
        bytesField:
          type: string
          format: byte
        binaryField:
          type: string
          format: binary
        dateField:
          type: string
          format: date
        dateTimeField:
          type: string
          format: date-time
        boolField:
          type: boolean
`,
			expected: `syntax = "proto3";

package testpkg;

message AllTypes {
  int32 int32_field = 1 [json_name = "int32Field"];
  int64 int64_field = 2 [json_name = "int64Field"];
  float float_field = 3 [json_name = "floatField"];
  double double_field = 4 [json_name = "doubleField"];
  string string_field = 5 [json_name = "stringField"];
  bytes bytes_field = 6 [json_name = "bytesField"];
  bytes binary_field = 7 [json_name = "binaryField"];
  string date_field = 8 [json_name = "dateField"];
  string date_time_field = 9 [json_name = "dateTimeField"];
  bool bool_field = 10 [json_name = "boolField"];
}
`,
		},
		{
			name: "default integer format",
			given: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas:
    Thing:
      type: object
      properties:
        count:
          type: integer
`,
			expected: `syntax = "proto3";

package testpkg;

message Thing {
  int32 count = 1;
}
`,
		},
		{
			name: "default number format",
			given: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas:
    Thing:
      type: object
      properties:
        value:
          type: number
`,
			expected: `syntax = "proto3";

package testpkg;

message Thing {
  double value = 1;
}
`,
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
  string email = 2;
  int32 age = 3;
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
  double amount = 3;
  string status = 4;
}
`

	result, err := conv.Convert([]byte(given), "testpkg")
	require.NoError(t, err)
	assert.Equal(t, expected, string(result))
}

func TestConvertJSONNameAnnotation(t *testing.T) {
	for _, test := range []struct {
		name     string
		given    string
		expected string
	}{
		{
			name: "userId gets json_name annotation",
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
`,
			expected: `syntax = "proto3";

package testpkg;

message User {
  string user_id = 1 [json_name = "userId"];
}
`,
		},
		{
			name: "email does not get json_name annotation",
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
        email:
          type: string
`,
			expected: `syntax = "proto3";

package testpkg;

message User {
  string email = 1;
}
`,
		},
		{
			name: "HTTPStatus gets json_name annotation (no acronym detection)",
			given: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas:
    Response:
      type: object
      properties:
        HTTPStatus:
          type: integer
`,
			expected: `syntax = "proto3";

package testpkg;

message Response {
  int32 h_t_t_p_status = 1 [json_name = "HTTPStatus"];
}
`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.given), "testpkg")
			require.NoError(t, err)
			assert.Equal(t, test.expected, string(result))
		})
	}
}
