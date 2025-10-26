package internal_test

import (
	"testing"

	conv "github.com/duh-rpc/openapi-proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
