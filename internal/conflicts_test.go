package internal_test

import (
	"testing"

	conv "github.com/duh-rpc/openapi-proto.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertNameConflicts(t *testing.T) {
	for _, test := range []struct {
		name     string
		given    string
		expected string
	}{
		{
			name: "duplicate message names get numeric suffixes",
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
        name:
          type: string
    user:
      type: object
      properties:
        email:
          type: string
    USER:
      type: object
      properties:
        id:
          type: integer
`,
			expected: `syntax = "proto3";

package testpkg;

message User {
  string name = 1 [json_name = "name"];
}

message User_2 {
  string email = 1 [json_name = "email"];
}

message User_3 {
  int32 id = 1 [json_name = "id"];
}
`,
		},
		{
			name: "duplicate enum names get numeric suffixes",
			given: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas:
    Status:
      type: string
      enum:
        - active
        - inactive
    status:
      type: string
      enum:
        - pending
        - completed
`,
			expected: `syntax = "proto3";

package testpkg;

enum Status {
  STATUS_UNSPECIFIED = 0;
  STATUS_ACTIVE = 1;
  STATUS_INACTIVE = 2;
}

enum Status_2 {
  STATUS_2_UNSPECIFIED = 0;
  STATUS_2_PENDING = 1;
  STATUS_2_COMPLETED = 2;
}
`,
		},
		{
			name: "mixed message and enum with same name",
			given: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas:
    Item:
      type: object
      properties:
        name:
          type: string
    item:
      type: string
      enum:
        - one
        - two
`,
			expected: `syntax = "proto3";

package testpkg;

message Item {
  string name = 1 [json_name = "name"];
}

enum Item_2 {
  ITEM_2_UNSPECIFIED = 0;
  ITEM_2_ONE = 1;
  ITEM_2_TWO = 2;
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
