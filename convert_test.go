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
  string userId = 1 [json_name = "userId"];
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
  string orderId = 1 [json_name = "orderId"];
  string customerId = 2 [json_name = "customerId"];
  double amount = 3 [json_name = "amount"];
  string status = 4 [json_name = "status"];
}
`

	result, err := conv.Convert([]byte(given), "testpkg")
	require.NoError(t, err)
	assert.Equal(t, expected, string(result))
}

func TestConvertCompleteExample(t *testing.T) {
	given := `openapi: 3.0.0
info:
  title: E-Commerce API
  version: 1.0.0
paths: {}
components:
  schemas:
    OrderStatus:
      type: string
      description: Status of an order
      enum:
        - pending
        - confirmed
        - shipped
        - delivered
        - cancelled

    Address:
      type: object
      description: Shipping or billing address
      properties:
        street:
          type: string
          description: Street address
        city:
          type: string
        state:
          type: string
        zipCode:
          type: string
        country:
          type: string

    Product:
      type: object
      description: Product in the catalog
      properties:
        productId:
          type: string
          description: Unique product identifier
        name:
          type: string
        description:
          type: string
        price:
          type: number
          format: double
        inStock:
          type: boolean
        category:
          type: string
          enum:
            - electronics
            - clothing
            - books
            - home

    OrderItem:
      type: object
      description: Item in an order
      properties:
        product:
          $ref: '#/components/schemas/Product'
        quantity:
          type: integer
          format: int32
        unitPrice:
          type: number
          format: double

    Order:
      type: object
      description: Customer order
      properties:
        orderId:
          type: string
          description: Unique order identifier
        customerId:
          type: string
        item:
          type: array
          items:
            $ref: '#/components/schemas/OrderItem'
        status:
          $ref: '#/components/schemas/OrderStatus'
        shippingAddress:
          $ref: '#/components/schemas/Address'
        totalAmount:
          type: number
          format: double
        createdAt:
          type: string
          format: date-time
`
	expected := `syntax = "proto3";

package ecommerce;

// Status of an order
enum OrderStatus {
  ORDER_STATUS_UNSPECIFIED = 0;
  ORDER_STATUS_PENDING = 1;
  ORDER_STATUS_CONFIRMED = 2;
  ORDER_STATUS_SHIPPED = 3;
  ORDER_STATUS_DELIVERED = 4;
  ORDER_STATUS_CANCELLED = 5;
}

// Shipping or billing address
message Address {
  // Street address
  string street = 1 [json_name = "street"];
  string city = 2 [json_name = "city"];
  string state = 3 [json_name = "state"];
  string zipCode = 4 [json_name = "zipCode"];
  string country = 5 [json_name = "country"];
}

enum Category {
  CATEGORY_UNSPECIFIED = 0;
  CATEGORY_ELECTRONICS = 1;
  CATEGORY_CLOTHING = 2;
  CATEGORY_BOOKS = 3;
  CATEGORY_HOME = 4;
}

// Product in the catalog
message Product {
  // Unique product identifier
  string productId = 1 [json_name = "productId"];
  string name = 2 [json_name = "name"];
  string description = 3 [json_name = "description"];
  double price = 4 [json_name = "price"];
  bool inStock = 5 [json_name = "inStock"];
  Category category = 6 [json_name = "category"];
}

// Item in an order
message OrderItem {
  Product product = 1 [json_name = "product"];
  int32 quantity = 2 [json_name = "quantity"];
  double unitPrice = 3 [json_name = "unitPrice"];
}

// Customer order
message Order {
  // Unique order identifier
  string orderId = 1 [json_name = "orderId"];
  string customerId = 2 [json_name = "customerId"];
  repeated OrderItem item = 3 [json_name = "item"];
  OrderStatus status = 4 [json_name = "status"];
  Address shippingAddress = 5 [json_name = "shippingAddress"];
  double totalAmount = 6 [json_name = "totalAmount"];
  string createdAt = 7 [json_name = "createdAt"];
}
`

	result, err := conv.Convert([]byte(given), "ecommerce")
	require.NoError(t, err)
	assert.Equal(t, expected, string(result))
}
