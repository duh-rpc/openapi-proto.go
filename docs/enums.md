# Enum Conversion

This document explains how the library handles OpenAPI enums, distinguishing between string enums and integer enums to preserve JSON wire format compatibility.

## Overview

The library handles two types of enums differently:

- **String enums** (type: string + enum) → `string` fields with enum values in comments
- **Integer enums** (type: integer + enum) → Protocol Buffer enum types

This approach ensures JSON wire format compatibility: string values remain strings in JSON, not integers or enum constant names.

## String Enums

String enums are mapped to `string` fields with enum values documented in comments. This preserves the original JSON wire format where enum values are transmitted as their literal string values.

### Basic Example

**OpenAPI Input:**
```yaml
components:
  schemas:
    Order:
      type: object
      properties:
        status:
          type: string
          description: Status of the order
          enum:
            - pending
            - confirmed
            - shipped
```

**Proto3 Output:**
```protobuf
message Order {
  // Status of the order
  // enum: [pending, confirmed, shipped]
  string status = 1 [json_name = "status"];
}
```

**JSON Wire Format:**
```json
{
  "status": "pending"
}
```

The JSON contains the exact string value `"pending"`, maintaining full compatibility with the OpenAPI specification.

### Referenced String Enums

Top-level string enum schemas referenced via `$ref` are handled the same way:

**OpenAPI Input:**
```yaml
components:
  schemas:
    OrderStatus:
      type: string
      description: Status of an order
      enum:
        - pending
        - confirmed
        - shipped

    Order:
      type: object
      properties:
        status:
          $ref: '#/components/schemas/OrderStatus'
```

**Proto3 Output:**
```protobuf
message Order {
  // Status of an order
  // enum: [pending, confirmed, shipped]
  string status = 1 [json_name = "status"];
}
```

Note that no standalone `OrderStatus` enum definition is generated - the enum values are documented directly on the field that references it.

### String Enums in Arrays

String enums in arrays become `repeated string` fields:

**OpenAPI Input:**
```yaml
components:
  schemas:
    Article:
      type: object
      properties:
        tags:
          type: array
          items:
            type: string
            enum:
              - draft
              - published
              - archived
```

**Proto3 Output:**
```protobuf
message Article {
  // enum: [draft, published, archived]
  repeated string tags = 1 [json_name = "tags"];
}
```

## Integer Enums

Integer enums are mapped to Protocol Buffer enum types, following proto3 conventions.

### Basic Example

**OpenAPI Input:**
```yaml
components:
  schemas:
    Code:
      type: integer
      enum:
        - 200
        - 400
        - 404
        - 500
```

**Proto3 Output:**
```protobuf
enum Code {
  CODE_UNSPECIFIED = 0;
  CODE_200 = 1;
  CODE_400 = 2;
  CODE_404 = 3;
  CODE_500 = 4;
}
```

### How It Works

1. **UNSPECIFIED Value**: Proto3 requires the first enum value to be `0`. An `<NAME>_UNSPECIFIED = 0` value is automatically added.

2. **Value Transformation**: Original enum values are converted to uppercase with the enum name as a prefix:
   - `200` → `CODE_200`
   - `404` → `CODE_404`

3. **Value Numbering**: Original values start at `1` and increment sequentially.

### Integer Enums in Messages

When a message field references an integer enum, the field description is cleared (not duplicated) since the description is hoisted to the enum definition:

**OpenAPI Input:**
```yaml
components:
  schemas:
    HttpCode:
      type: integer
      description: Standard HTTP status codes
      enum:
        - 200
        - 404
        - 500

    Response:
      type: object
      properties:
        code:
          $ref: '#/components/schemas/HttpCode'
```

**Proto3 Output:**
```protobuf
// Standard HTTP status codes
enum HttpCode {
  HTTP_CODE_UNSPECIFIED = 0;
  HTTP_CODE_200 = 1;
  HTTP_CODE_404 = 2;
  HTTP_CODE_500 = 3;
}

message Response {
  HttpCode code = 1 [json_name = "code"];
}
```

## Wire Format Compatibility

The key difference between string and integer enums is JSON wire format:

### String Enum Wire Format

String enums preserve exact string values in JSON:

```json
{
  "status": "pending"
}
```

This matches the OpenAPI specification exactly. The protobuf field is `string`, so JSON serialization uses the actual string value.

### Integer Enum Wire Format

Integer enums use proto3's standard enum JSON encoding:

```json
{
  "code": "CODE_200"
}
```

The JSON contains the enum constant name (e.g., `CODE_200`), not the original integer value (`200`). This is proto3's standard behavior for enum fields.

## When to Use Each Type

### Use String Enums When:

- You need exact JSON wire format compatibility with an existing API
- The enum represents categorical string values (e.g., statuses, types, categories)
- You're designing a new API and prefer string-based enums for readability
- External systems expect literal string values in JSON

### Use Integer Enums When:

- The enum represents numeric codes (HTTP status codes, error codes)
- You want protobuf's type safety for enum values
- You're okay with enum constant names in JSON (e.g., `"CODE_200"` instead of `200`)

## Validation Rules

The converter enforces validation rules for enum schemas to prevent ambiguity and ensure correctness:

### Required: Explicit Type Field

All enum schemas must have an explicit `type` field (`string` or `integer`). Type inference is not supported.

**Valid:**
```yaml
Status:
  type: string
  enum: [active, inactive]
```

**Invalid (will error):**
```yaml
Status:
  enum: [active, inactive]  # Error: enum must have explicit type field
```

### Rejected: Null Values

Enums cannot contain null values.

**Valid:**
```yaml
Status:
  type: string
  enum: [active, inactive, unknown]
```

**Invalid (will error):**
```yaml
Status:
  type: string
  enum: [active, inactive, null]  # Error: enum cannot contain null values
```

### Rejected: Mixed Types

All enum values must be the same type (all strings or all integers).

**Valid:**
```yaml
Code:
  type: integer
  enum: [200, 404, 500]
```

**Invalid (will error):**
```yaml
Code:
  type: integer
  enum: [200, "404", 500]  # Error: enum contains mixed types (string and integer)
```

### Allowed: Empty Enums

Empty enum arrays are allowed and result in no enum comment being generated:

**Valid:**
```yaml
Status:
  type: string
  enum: []
```

**Generated:**
```protobuf
message Order {
  string status = 1 [json_name = "status"];
}
```

No `// enum: []` comment is generated for empty enums.

### Allowed: Duplicate Values

Duplicate enum values are allowed (no deduplication is performed):

**Valid:**
```yaml
Status:
  type: string
  enum: [active, active, inactive]
```

**Generated:**
```protobuf
message Order {
  // enum: [active, active, inactive]
  string status = 1 [json_name = "status"];
}
```

### Case Sensitivity

Enum values are case-sensitive. `"Active"` and `"active"` are treated as distinct values:

**Valid:**
```yaml
Status:
  type: string
  enum: [Active, active, ACTIVE]
```

All three values are preserved as distinct.

## Special Characters in String Enums

String enum values can contain special characters, which are preserved in the comment:

**OpenAPI Input:**
```yaml
components:
  schemas:
    Tag:
      type: string
      enum:
        - foo bar
        - a"b
        - c[d]
```

**Proto3 Output:**
```protobuf
message Article {
  // enum: [foo bar, a"b, c[d]]
  string tag = 1 [json_name = "tag"];
}
```

Values with spaces, quotes, or brackets require no escaping in proto comments.

## Linter Integration

The validation rules documented above are enforced by the converter at build time. External linters can implement the same validation rules to provide earlier feedback during development:

- Enum must have explicit type field (string or integer)
- Enum cannot contain null values
- Enum cannot contain mixed types
- Empty enum arrays are allowed
- Duplicate enum values are allowed
- Enum values are case-sensitive

By implementing these rules in a linter, you can catch enum schema issues before running the converter.

## Migration from Previous Versions

**Breaking Change:** Previous versions of this library converted all enums (both string and integer) to protobuf enum types. This broke JSON wire format compatibility for string enums.

**Before (all enums → protobuf enums):**
```yaml
Status:
  type: string
  enum: [active, inactive]
```

Generated:
```protobuf
enum Status {
  STATUS_UNSPECIFIED = 0;
  STATUS_ACTIVE = 1;
  STATUS_INACTIVE = 2;
}
```

JSON wire format:
```json
{"status": "STATUS_ACTIVE"}  // ❌ Wrong - should be "active"
```

**After (string enums → string fields):**
```yaml
Status:
  type: string
  enum: [active, inactive]
```

Generated:
```protobuf
message User {
  // enum: [active, inactive]
  string status = 1 [json_name = "status"];
}
```

JSON wire format:
```json
{"status": "active"}  // ✅ Correct - exact match
```

If you have existing code that relies on the old behavior, you'll need to:

1. Update any code expecting protobuf enum types for string enums
2. Verify JSON serialization now uses string values instead of enum constant names
3. Update tests expecting enum constant names in JSON

## Best Practices

1. **Use explicit types**: Always specify `type: string` or `type: integer` for enums, never rely on type inference.

2. **Choose the right enum type**:
   - Use string enums for categorical values that should be human-readable in JSON
   - Use integer enums for numeric codes where type safety is more important than readability

3. **Document enum values**: Add descriptions to enum schemas - for string enums, the description appears with the enum comment.

4. **Consider external APIs**: If you're building an API consumed by external systems, string enums provide better compatibility since the JSON contains literal values.

5. **Validate early**: Use linters to validate enum schemas during development rather than waiting for build-time errors.

## Further Reading

- [Protocol Buffers Enum Documentation](https://protobuf.dev/programming-guides/proto3/#enum)
- [Proto3 JSON Mapping](https://protobuf.dev/programming-guides/proto3/#json)
- [OpenAPI Enum Specification](https://swagger.io/specification/#schema-object)
