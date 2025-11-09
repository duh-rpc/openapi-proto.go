package internal_test

import (
	"testing"

	conv "github.com/duh-rpc/openapi-proto.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTopLevelEnum(t *testing.T) {
	given := `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
components:
  schemas:
    Status:
      type: string
      enum:
        - active
        - inactive
        - pending
        - isActive`

	expected := `syntax = "proto3";

package testpkg;

option go_package = "github.com/example/proto/v1";

enum Status {
  STATUS_UNSPECIFIED = 0;
  STATUS_ACTIVE = 1;
  STATUS_INACTIVE = 2;
  STATUS_PENDING = 3;
  STATUS_IS_ACTIVE = 4;
}

`

	result, err := conv.Convert([]byte(given), conv.ConvertOptions{
		PackageName: "testpkg",
		PackagePath: "github.com/example/proto/v1",
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, expected, string(result.Protobuf))
}

func TestEnumWithDashes(t *testing.T) {
	given := `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
components:
  schemas:
    Status:
      type: string
      enum:
        - in-progress
        - not-started
        - completed`

	expected := `syntax = "proto3";

package testpkg;

option go_package = "github.com/example/proto/v1";

enum Status {
  STATUS_UNSPECIFIED = 0;
  STATUS_IN_PROGRESS = 1;
  STATUS_NOT_STARTED = 2;
  STATUS_COMPLETED = 3;
}

`

	result, err := conv.Convert([]byte(given), conv.ConvertOptions{
		PackageName: "testpkg",
		PackagePath: "github.com/example/proto/v1",
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, expected, string(result.Protobuf))
}

func TestEnumWithNumbers(t *testing.T) {
	given := `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
components:
  schemas:
    Code:
      type: integer
      enum:
        - 200
        - 401
        - 404
        - 500`

	expected := `syntax = "proto3";

package testpkg;

option go_package = "github.com/example/proto/v1";

enum Code {
  CODE_UNSPECIFIED = 0;
  CODE_200 = 1;
  CODE_401 = 2;
  CODE_404 = 3;
  CODE_500 = 4;
}

`

	result, err := conv.Convert([]byte(given), conv.ConvertOptions{
		PackageName: "testpkg",
		PackagePath: "github.com/example/proto/v1",
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, expected, string(result.Protobuf))
}

func TestEnumWithDescription(t *testing.T) {
	given := `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
components:
  schemas:
    Status:
      type: string
      description: Status of the operation
      enum:
        - active
        - inactive`

	expected := `syntax = "proto3";

package testpkg;

option go_package = "github.com/example/proto/v1";

// Status of the operation
enum Status {
  STATUS_UNSPECIFIED = 0;
  STATUS_ACTIVE = 1;
  STATUS_INACTIVE = 2;
}

`

	result, err := conv.Convert([]byte(given), conv.ConvertOptions{
		PackageName: "testpkg",
		PackagePath: "github.com/example/proto/v1",
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, expected, string(result.Protobuf))
}

func TestInlineEnum(t *testing.T) {
	given := `openapi: 3.0.0
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
        status:
          type: string
          enum:
            - active
            - inactive
            - notStarted`

	expected := `syntax = "proto3";

package testpkg;

option go_package = "github.com/example/proto/v1";

enum Status {
  STATUS_UNSPECIFIED = 0;
  STATUS_ACTIVE = 1;
  STATUS_INACTIVE = 2;
  STATUS_NOT_STARTED = 3;
}

message User {
  string name = 1 [json_name = "name"];
  Status status = 2 [json_name = "status"];
}

`

	result, err := conv.Convert([]byte(given), conv.ConvertOptions{
		PackageName: "testpkg",
		PackagePath: "github.com/example/proto/v1",
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, expected, string(result.Protobuf))
}

func TestMultipleInlineEnums(t *testing.T) {
	given := `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
components:
  schemas:
    User:
      type: object
      properties:
        status:
          type: string
          enum:
            - active
            - inactive
        role:
          type: string
          enum:
            - admin
            - user
            - superAdmin`

	expected := `syntax = "proto3";

package testpkg;

option go_package = "github.com/example/proto/v1";

enum Status {
  STATUS_UNSPECIFIED = 0;
  STATUS_ACTIVE = 1;
  STATUS_INACTIVE = 2;
}

enum Role {
  ROLE_UNSPECIFIED = 0;
  ROLE_ADMIN = 1;
  ROLE_USER = 2;
  ROLE_SUPER_ADMIN = 3;
}

message User {
  Status status = 1 [json_name = "status"];
  Role role = 2 [json_name = "role"];
}

`

	result, err := conv.Convert([]byte(given), conv.ConvertOptions{
		PackageName: "testpkg",
		PackagePath: "github.com/example/proto/v1",
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, expected, string(result.Protobuf))
}

func TestEnumAndMessageMixed(t *testing.T) {
	given := `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
components:
  schemas:
    Status:
      type: string
      enum:
        - active
        - inactive
    User:
      type: object
      properties:
        name:
          type: string
    Priority:
      type: string
      enum:
        - high
        - low`

	expected := `syntax = "proto3";

package testpkg;

option go_package = "github.com/example/proto/v1";

enum Status {
  STATUS_UNSPECIFIED = 0;
  STATUS_ACTIVE = 1;
  STATUS_INACTIVE = 2;
}

message User {
  string name = 1 [json_name = "name"];
}

enum Priority {
  PRIORITY_UNSPECIFIED = 0;
  PRIORITY_HIGH = 1;
  PRIORITY_LOW = 2;
}

`

	result, err := conv.Convert([]byte(given), conv.ConvertOptions{
		PackageName: "testpkg",
		PackagePath: "github.com/example/proto/v1",
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, expected, string(result.Protobuf))
}
