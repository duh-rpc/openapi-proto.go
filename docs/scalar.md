# Scalar Type Conversion

This document explains how OpenAPI scalar types are converted to Protocol Buffer 3 (proto3) types, including field naming conventions and JSON serialization behavior.

## Overview

Scalar types are the basic data types in OpenAPI specifications (strings, numbers, integers, booleans). This library converts these types to their proto3 equivalents while preserving the original OpenAPI field names for JSON serialization compatibility.

## Type Mapping Table

The following table shows how OpenAPI types and formats map to proto3 scalar types:

| OpenAPI Type | OpenAPI Format | Proto3 Type | Notes |
|--------------|----------------|-------------|-------|
| `integer` | (none or `int32`) | `int32` | Default integer format |
| `integer` | `int64` | `int64` | 64-bit integers |
| `number` | `float` | `float` | 32-bit floating point |
| `number` | (none or `double`) | `double` | Default number format |
| `string` | (none) | `string` | Text strings |
| `string` | `byte` or `binary` | `bytes` | Binary data |
| `string` | `date` or `date-time` | `string` | Dates stored as strings |
| `boolean` | (any) | `bool` | Boolean values |

### Notes on Type Mapping

- **Default Integer**: When no `format` is specified for `integer` types, `int32` is used
- **Default Number**: When no `format` is specified for `number` types, `double` is used
- **Date/DateTime**: These are converted to `string` rather than using `google.protobuf.Timestamp` for simplicity
- **Binary Data**: Both `byte` (base64-encoded) and `binary` formats map to proto3 `bytes`

## Field Naming Convention

### Proto Field Names: Preserved from OpenAPI

Field names are preserved from the OpenAPI schema when they meet proto3 syntax requirements:

```yaml
# OpenAPI
properties:
  userId:
    type: string
  HTTPStatus:
    type: integer
  email:
    type: string
```

```protobuf
// Proto3
message User {
  string userId = 1 [json_name = "userId"];
  int32 HTTPStatus = 2 [json_name = "HTTPStatus"];
  string email = 3 [json_name = "email"];
}
```

### Field Name Sanitization

Invalid characters are replaced with underscores to meet proto3 syntax requirements:
- Hyphens, dots, and spaces become underscores
- Field names must start with an ASCII letter (A-Z, a-z)
- Field names can contain ASCII letters, digits (0-9), and underscores (_)

Examples:
- `userId` → `userId` (preserved)
- `HTTPStatus` → `HTTPStatus` (preserved)
- `user_id` → `user_id` (preserved)
- `status-code` → `status_code` (hyphen replaced)
- `user.name` → `user_name` (dot replaced)

### JSON Name Annotations: Always Present

**All fields include a `json_name` annotation** that preserves the original OpenAPI field name. This ensures:
1. **Consistency**: Every field explicitly states its JSON serialization name
2. **Clarity**: No ambiguity about what the JSON key will be
3. **Compatibility**: Original OpenAPI naming convention is preserved

```protobuf
message User {
  string userId = 1 [json_name = "userId"];       // Preserved camelCase
  string email = 2 [json_name = "email"];         // Explicit even when simple
  string user_name = 3 [json_name = "user_name"]; // Preserved snake_case
}
```

### Why json_name Is Always Included

The `json_name` annotation is always included for all fields. This provides:

- **Explicit documentation** of JSON serialization behavior
- **Consistent proto file structure** (all fields have the same format)
- **Future-proof design** (changes to naming logic won't break existing protos)

## Protoc Go Code Generation

When the proto3 file is processed by `protoc` to generate Go code, the following transformations occur:

### Proto Field to Go Struct Field

The protoc compiler automatically converts proto field names to PascalCase Go struct field names:

```protobuf
// Proto3
message User {
  string userId = 1 [json_name = "userId"];
  int32 age = 2 [json_name = "age"];
}
```

```go
// Generated Go Code
type User struct {
	UserId string `protobuf:"bytes,1,opt,name=userId,json=userId" json:"userId,omitempty"`
	Age    int32  `protobuf:"varint,2,opt,name=age,json=age" json:"age,omitempty"`
}
```

**Key Points:**
- Proto field `userId` becomes Go struct field `UserId` (exported, PascalCase)
- JSON marshaling uses the `json_name` value: `userId` for the first field, `age` for the second
- The `protobuf` struct tag preserves the original proto field name

### JSON Serialization Behavior

When Go structs are marshaled to/from JSON using `protojson`:

```go
user := &User{
	UserId: "123",
	Age:    25,
}

// JSON output (uses json_name annotations)
{
  "userId": "123",  // Uses json_name from proto
  "age": 25         // Uses json_name from proto
}
```

The `json_name` annotation in the proto file determines the JSON key, **not** the Go struct field name or proto field name.

## Complete Example

### OpenAPI Schema
```yaml
openapi: 3.0.0
info:
  title: User API
  version: 1.0.0
components:
  schemas:
    User:
      type: object
      properties:
        userId:           # camelCase
          type: string
        emailAddress:     # camelCase
          type: string
        age:              # lowercase
          type: integer
        HTTPStatus:       # PascalCase with acronym
          type: integer
        created_at:       # snake_case
          type: string
          format: date-time
```

### Generated Proto3
```protobuf
syntax = "proto3";

package userapi;

message User {
  string userId = 1 [json_name = "userId"];
  string emailAddress = 2 [json_name = "emailAddress"];
  int32 age = 3 [json_name = "age"];
  int32 HTTPStatus = 4 [json_name = "HTTPStatus"];
  string created_at = 5 [json_name = "created_at"];
}
```

### Generated Go Code
```go
type User struct {
	UserId        string `protobuf:"bytes,1,opt,name=userId,json=userId" json:"userId,omitempty"`
	EmailAddress  string `protobuf:"bytes,2,opt,name=emailAddress,json=emailAddress" json:"emailAddress,omitempty"`
	Age           int32  `protobuf:"varint,3,opt,name=age,json=age" json:"age,omitempty"`
	HTTPStatus    int32  `protobuf:"varint,4,opt,name=HTTPStatus,json=HTTPStatus" json:"HTTPStatus,omitempty"`
	CreatedAt     string `protobuf:"bytes,5,opt,name=created_at,json=created_at" json:"created_at,omitempty"`
}
```

### JSON Serialization
```json
{
  "userId": "abc123",
  "emailAddress": "user@example.com",
  "age": 30,
  "HTTPStatus": 200,
  "created_at": "2025-01-15T10:30:00Z"
}
```

Notice how:
- Proto fields preserve original names: `userId`, `emailAddress`, `HTTPStatus`, `created_at`
- Go struct fields are all PascalCase: `UserId`, `EmailAddress`, `HTTPStatus`, `CreatedAt`
- JSON keys match the original OpenAPI names via `json_name` annotations

## Best Practices

### For OpenAPI Schemas

1. **Be consistent** with naming conventions in your OpenAPI specs
2. **Avoid plural property names** for inline objects or enums in arrays (use `$ref` instead)
3. **Document naming choices** in your OpenAPI description fields

### For Proto Generation

1. **Trust the toolchain**: Let protoc handle Go struct field naming
2. **Review generated protos**: Ensure json_name annotations match your API contract
3. **Test JSON serialization**: Verify that generated Go code produces the expected JSON format

## Related Documentation

- [Array Type Conversion](./arrays.md) - How arrays are converted to repeated fields
- [Enum Conversion](./enums.md) - How enums are hoisted and named
- [Nested Objects](./nested.md) - How inline objects become nested messages

## Troubleshooting

### Why is my field named differently in Go?

Protoc automatically converts proto fields to PascalCase Go fields. This is standard protoc behavior and cannot be changed.

### Why does my JSON use different naming than expected?

Check the `json_name` annotation in your proto file. The JSON key comes from `json_name`, which preserves the original OpenAPI field name.

### What about acronyms like HTTP, ID, URL?

The converter preserves acronyms as-is from your OpenAPI schema. If your OpenAPI field is `HTTPStatus`, it will be `HTTPStatus` in the proto file and `HTTPStatus` in Go.
