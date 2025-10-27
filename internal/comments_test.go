package internal_test

import (
	"testing"

	conv "github.com/duh-rpc/openapi-proto.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDescriptionComments(t *testing.T) {
	for _, test := range []struct {
		name     string
		given    string
		expected string
	}{
		{
			name: "schema with description",
			given: `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
components:
  schemas:
    User:
      type: object
      description: Represents a user in the system
      properties:
        name:
          type: string
`,
			expected: `syntax = "proto3";

package testpkg;

option go_package = "github.com/example/proto/v1";

// Represents a user in the system
message User {
  string name = 1 [json_name = "name"];
}

`,
		},
		{
			name: "field with description",
			given: `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
components:
  schemas:
    User:
      type: object
      properties:
        email:
          type: string
          description: User's email address
`,
			expected: `syntax = "proto3";

package testpkg;

option go_package = "github.com/example/proto/v1";

message User {
  // User's email address
  string email = 1 [json_name = "email"];
}

`,
		},
		{
			name: "no description",
			given: `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
components:
  schemas:
    User:
      type: object
      properties:
        name:
          type: string
`,
			expected: `syntax = "proto3";

package testpkg;

option go_package = "github.com/example/proto/v1";

message User {
  string name = 1 [json_name = "name"];
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
			assert.Equal(t, test.expected, string(result))
		})
	}
}

func TestMultiLineDescription(t *testing.T) {
	given := `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
components:
  schemas:
    User:
      type: object
      description: |-
        A user object that contains personal information.
        This includes name, email, and contact details.
        Used across the entire application.
      properties:
        name:
          type: string
          description: |-
            The full name of the user.
            Can include middle names and suffixes.
`

	expected := `syntax = "proto3";

package testpkg;

option go_package = "github.com/example/proto/v1";

// A user object that contains personal information.
// This includes name, email, and contact details.
// Used across the entire application.
message User {
  // The full name of the user.
  // Can include middle names and suffixes.
  string name = 1 [json_name = "name"];
}

`

	result, err := conv.Convert([]byte(given), conv.ConvertOptions{
		PackageName: "testpkg",
		PackagePath: "github.com/example/proto/v1",
	})
	require.NoError(t, err)
	assert.Equal(t, expected, string(result))
}

func TestBlankLineInDescription(t *testing.T) {
	given := `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
components:
  schemas:
    User:
      type: object
      description: |-
        First paragraph of description.

        Second paragraph after blank line.
      properties:
        email:
          type: string
`

	expected := `syntax = "proto3";

package testpkg;

option go_package = "github.com/example/proto/v1";

// First paragraph of description.
//
// Second paragraph after blank line.
message User {
  string email = 1 [json_name = "email"];
}

`

	result, err := conv.Convert([]byte(given), conv.ConvertOptions{
		PackageName: "testpkg",
		PackagePath: "github.com/example/proto/v1",
	})
	require.NoError(t, err)
	assert.Equal(t, expected, string(result))
}
