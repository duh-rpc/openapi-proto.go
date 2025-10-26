# OpenAPI to Protobuf Converter

A Go library that converts OpenAPI 3.0 schema definitions to Protocol Buffer 3 (proto3) format.

## Overview

This library parses OpenAPI 3.0 specifications and generates corresponding `.proto` files with proper type mappings, JSON field name annotations, and protobuf conventions. It's designed for projects that need to support both OpenAPI and protobuf interfaces.

## Installation

```bash
go get github.com/duh-rpc/openapi-proto
```

## Usage

### Basic Example

```go
package main

import (
    "fmt"
    "os"

    conv "github.com/duh-rpc/openapi-proto"
)

func main() {
    // Read OpenAPI specification
    openapi, err := os.ReadFile("api.yaml")
    if err != nil {
        panic(err)
    }

    // Convert to proto3
    proto, err := conv.Convert(openapi, "myapi")
    if err != nil {
        panic(err)
    }

    // Write proto file
    err = os.WriteFile("api.proto", proto, 0644)
    if err != nil {
        panic(err)
    }
}
```

### Input: OpenAPI 3.0 YAML

```yaml
openapi: 3.0.0
info:
  title: User API
  version: 1.0.0
paths: {}
components:
  schemas:
    User:
      type: object
      description: A user account
      properties:
        userId:
          type: string
          description: Unique user identifier
        email:
          type: string
        age:
          type: integer
        isActive:
          type: boolean
```

### Output: Proto3

```protobuf
syntax = "proto3";

package myapi;

// A user account
message User {
  // Unique user identifier
  string user_id = 1 [json_name = "userId"];
  string email = 2 [json_name = "email"];
  int32 age = 3 [json_name = "age"];
  bool is_active = 4 [json_name = "isActive"];
}
```

## Supported Features

### OpenAPI Features
- ✅ Object schemas with properties
- ✅ Scalar types (string, integer, number, boolean)
- ✅ Enums (top-level and inline)
- ✅ Arrays (repeated fields)
- ✅ Nested objects
- ✅ Schema references (`$ref`)
- ✅ Descriptions (converted to comments)
- ✅ Multiple format specifiers (int32, int64, float, double, byte, binary, date, date-time)

### Proto3 Features
- ✅ Message definitions
- ✅ Enum definitions with UNSPECIFIED values
- ✅ Repeated fields
- ✅ Nested messages
- ✅ JSON name annotations
- ✅ Field numbering (sequential based on YAML order)
- ✅ Comments from descriptions

## Unsupported Features

### OpenAPI Features Not Supported
- ❌ Schema composition: `allOf`, `anyOf`, `oneOf`, `not`
- ❌ External file references (only internal `#/components/schemas` refs)
- ❌ Nested arrays (e.g., `array` of `array`)
- ❌ Multi-type properties (e.g., `type: [string, null]`)
- ❌ Map types via `additionalProperties`
- ❌ Validation constraints (min, max, pattern, etc. are ignored)
- ❌ Polymorphism and discriminators
- ❌ OpenAPI 2.0 (Swagger) - only 3.x supported

### Proto3 Features Not Generated
- ❌ Service definitions
- ❌ Multiple output files (single file only)
- ❌ Import statements
- ❌ Proto options beyond `json_name`
- ❌ Map types
- ❌ `optional` keyword (all fields follow proto3 default semantics)
- ❌ Wrapper types for nullable fields

### Ignored OpenAPI Directives
- The `required` array is ignored (proto3 has no required keyword)
- The `nullable` field is ignored (proto3 uses zero values)

## Type Mapping

| OpenAPI Type | OpenAPI Format | Proto3 Type |
|--------------|----------------|-------------|
| string       | (none)         | string      |
| string       | byte           | bytes       |
| string       | binary         | bytes       |
| string       | date           | string      |
| string       | date-time      | string      |
| integer      | (none)         | int32       |
| integer      | int32          | int32       |
| integer      | int64          | int64       |
| number       | (none)         | double      |
| number       | float          | float       |
| number       | double         | double      |
| boolean      | (any)          | bool        |
| object       | (any)          | message     |
| array        | (any)          | repeated    |

## Naming Conventions

### Field Names: camelCase → snake_case

The library uses a simple letter-by-letter algorithm to convert field names from camelCase to snake_case:

- Each uppercase letter becomes lowercase with an underscore prefix (except the first character)
- **No special acronym detection** - each letter is handled individually

Examples:
- `userId` → `user_id`
- `HTTPStatus` → `h_t_t_p_status` (not `http_status`)
- `http2Protocol` → `http2_protocol`
- `email` → `email` (no conversion needed)

All fields include a `json_name` annotation to preserve the original OpenAPI field name for JSON serialization.

### Message Names: PascalCase

Schema names and nested message names are converted to PascalCase:
- `user_account` → `UserAccount`
- `shippingAddress` → `ShippingAddress`

### Enum Values: UPPERCASE_SNAKE_CASE

Enum values are prefixed with the enum name and converted to uppercase:
- Enum `Status` with value `active` → `STATUS_ACTIVE`
- Enum `Status` with value `in-progress` → `STATUS_IN_PROGRESS`
- Enum `Code` with value `401` → `CODE_401`

All enums automatically include an `UNSPECIFIED` value at position 0 following proto3 conventions.

### Plural Name Validation

When using inline objects or enums in arrays, property names **must be singular**:

```yaml
# ✅ GOOD - singular property name
properties:
  contact:
    type: array
    items:
      type: object
      properties:
        name:
          type: string

# ❌ BAD - plural property name
properties:
  contacts:  # Will cause error
    type: array
    items:
      type: object
```

**Why?** The library derives message names from property names. A plural property name like `contacts` would generate a message named `Contacts`, which is confusing. Instead:
- Use singular names: `contact` → `Contact` message
- Or use `$ref` to reference a named schema


## Examples

### Enums

**OpenAPI:**
```yaml
components:
  schemas:
    Status:
      type: string
      enum:
        - active
        - pending
        - inactive
```

**Proto3:**
```protobuf
enum Status {
  STATUS_UNSPECIFIED = 0;
  STATUS_ACTIVE = 1;
  STATUS_PENDING = 2;
  STATUS_INACTIVE = 3;
}
```

### Nested Objects

**OpenAPI:**
```yaml
components:
  schemas:
    User:
      type: object
      properties:
        name:
          type: string
        address:
          type: object
          properties:
            street:
              type: string
            city:
              type: string
```

**Proto3:**
```protobuf
message User {
  message Address {
    string street = 1 [json_name = "street"];
    string city = 2 [json_name = "city"];
  }

  string name = 1 [json_name = "name"];
  Address address = 2 [json_name = "address"];
}
```

### Arrays with References

**OpenAPI:**
```yaml
components:
  schemas:
    Address:
      type: object
      properties:
        street:
          type: string
        city:
          type: string

    User:
      type: object
      properties:
        name:
          type: string
        address:
          type: array
          items:
            $ref: '#/components/schemas/Address'
```

**Proto3:**
```protobuf
message Address {
  string street = 1 [json_name = "street"];
  string city = 2 [json_name = "city"];
}

message User {
  string name = 1 [json_name = "name"];
  repeated Address address = 2 [json_name = "address"];
}
```

### Name Conflict Resolution

When multiple schemas have the same name, numeric suffixes are automatically added:

**OpenAPI:**
```yaml
components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: string

    User:  # Duplicate name
      type: object
      properties:
        name:
          type: string
```

**Proto3:**
```protobuf
message User {
  string id = 1 [json_name = "id"];
}

message User_2 {
  string name = 1 [json_name = "name"];
}
```

## Best Practices

1. **Use singular property names** for arrays with inline objects/enums, or use `$ref` to reference named schemas
2. **Avoid acronyms in field names** if you need specific casing in proto3 (e.g., prefer `httpStatus` over `HTTPStatus`)
3. **Use descriptions** liberally - they become useful comments in the generated proto
4. **Order schemas intentionally** in your OpenAPI YAML - the output order will match
5. **Test with protoc** after generation to catch any proto3 reserved keywords

## Development

### Running Tests

```bash
make test
```

### Test Coverage

```bash
make coverage
```

### Linting

```bash
make lint
```

### Detailed Documentation
See the following links for more details:
- Detailed Enum Docs [Enums](docs/enums.md)
- Detailed Scalar Docs [Scalar](docs/scalar.md)
- Detailed Objects Docs [Objects](docs/objects.md)

## License

MIT License - see LICENSE file for details

## Related Projects

- [pb33f/libopenapi](https://github.com/pb33f/libopenapi) - OpenAPI parser used by this library
- [protocolbuffers/protobuf](https://github.com/protocolbuffers/protobuf) - Protocol Buffers

## Acknowledgments

This library uses the excellent [libopenapi](https://github.com/pb33f/libopenapi) for OpenAPI parsing, which provides
 support for OpenAPI 3.0 and 3.1 specifications.
