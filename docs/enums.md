# Enum Conversion

This document explains how OpenAPI string enums are converted to Protocol Buffer 3 enums, and the important limitations you need to understand.

## Overview

Protocol Buffers (proto3) only supports **integer enums**, while OpenAPI supports **string enums**. When this library converts an OpenAPI specification to proto3, string enum values are transformed into uppercase enum constant names.

## Conversion Example

### OpenAPI Input

```yaml
components:
  schemas:
    Status:
      type: string
      enum:
        - active
        - inactive
        - pending
```

### Proto3 Output

```proto
enum Status {
  STATUS_UNSPECIFIED = 0;
  STATUS_ACTIVE = 1;
  STATUS_INACTIVE = 2;
  STATUS_PENDING = 3;
}
```

## How It Works

1. **UNSPECIFIED Value**: Proto3 requires the first enum value to be `0`. We automatically add `<NAME>_UNSPECIFIED = 0`.

2. **Value Transformation**: Original enum values are converted to uppercase with the enum name as a prefix:
   - `"active"` → `STATUS_ACTIVE`
   - `"in-progress"` → `STATUS_IN_PROGRESS` (dashes converted to underscores)
   - `"404"` → `STATUS_404` (numbers preserved)

3. **Value Numbering**: Original values start at `1` and increment sequentially.

## Important Limitation: Original String Values Are Not Preserved

**The original OpenAPI string values are lost during conversion.** Protocol Buffers enums are fundamentally integer types with symbolic names, not string types.

### What You Get When Unmarshalling

#### From Binary Protobuf

```go
import pb "your/generated/proto/package"

var msg pb.MyMessage
proto.Unmarshal(data, &msg)

// msg.Status is an integer enum value
switch msg.Status {
case pb.Status_STATUS_ACTIVE:
    fmt.Println("Status is active")
case pb.Status_STATUS_INACTIVE:
    fmt.Println("Status is inactive")
}

// Get the enum constant name as a string
name := msg.Status.String()  // Returns "STATUS_ACTIVE"
```

#### From JSON (Proto3 JSON Mapping)

When using proto3's JSON encoding, enums are represented as strings containing the **enum constant name**, not the original OpenAPI value:

```json
{
  "status": "STATUS_ACTIVE"
}
```

**Not** the original:
```json
{
  "status": "active"
}
```

## Workarounds

If you need to recover the original lowercase string values from OpenAPI, here are three approaches:

### Option 1: String Conversion Function

Transform the enum constant name back to the original format:

```go
func StatusToOriginalValue(status pb.Status) string {
    name := status.String()                    // "STATUS_ACTIVE"
    name = strings.TrimPrefix(name, "STATUS_") // "ACTIVE"
    return strings.ToLower(name)               // "active"
}

// Usage
originalValue := StatusToOriginalValue(msg.Status)  // "active"
```

**Caveat**: This won't work correctly for enums with dashes:
- `STATUS_IN_PROGRESS` → `"in_progress"` (not `"in-progress"`)

### Option 2: Maintain a Mapping

Create a bidirectional mapping in your code:

```go
var statusToString = map[pb.Status]string{
    pb.Status_STATUS_ACTIVE:   "active",
    pb.Status_STATUS_INACTIVE: "inactive",
    pb.Status_STATUS_PENDING:  "pending",
}

var stringToStatus = map[string]pb.Status{
    "active":   pb.Status_STATUS_ACTIVE,
    "inactive": pb.Status_STATUS_INACTIVE,
    "pending":  pb.Status_STATUS_PENDING,
}

// Usage
originalValue := statusToString[msg.Status]  // "active"
enumValue := stringToStatus["active"]        // pb.Status_STATUS_ACTIVE
```

**Pros**: Exact control over string values, handles dashes correctly.

**Cons**: Requires manual maintenance when enums change.

### Option 3: Use String Fields Instead

If you need exact string preservation and don't need the type safety of enums, define the field as a string in your OpenAPI spec without enum constraints:

```yaml
components:
  schemas:
    User:
      type: object
      properties:
        status:
          type: string
          description: "Status values: active, inactive, pending"
```

This generates:

```proto
message User {
  string status = 1;
}
```

**Pros**: Original string values preserved exactly.

**Cons**: No type safety, no compile-time validation.

## Inline Enums

Inline enums defined within object properties are automatically hoisted to top-level enum definitions:

### OpenAPI Input

```yaml
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
```

### Proto3 Output

```proto
enum Status {
  STATUS_UNSPECIFIED = 0;
  STATUS_ACTIVE = 1;
  STATUS_INACTIVE = 2;
}

message User {
  string name = 1;
  Status status = 2;
}
```

The enum is named after the property name (`status` → `Status`).

## Enum Value Naming Rules

The library applies these transformations to enum values:

1. Convert to UPPERCASE
2. Replace dashes (`-`) with underscores (`_`)
3. Preserve numbers as-is
4. Prefix with `<ENUM_NAME>_`

### Examples

| OpenAPI Value | Proto3 Constant      |
|---------------|----------------------|
| `active`      | `STATUS_ACTIVE`      |
| `in-progress` | `STATUS_IN_PROGRESS` |
| `404`         | `CODE_404`           |
| `not-found`   | `ERROR_NOT_FOUND`    |

## Best Practices

1. **Document the limitation**: Make sure your team understands that enum constant names, not original values, are used in JSON serialization.

2. **Consider your use case**:
   - If you need exact string preservation for external APIs, consider using string fields instead of enums.
   - If type safety is more important than exact string values, use enums and apply one of the workarounds.

3. **Naming conventions**: Use clear, descriptive enum values in your OpenAPI spec since they become the basis for proto3 constant names:
   - Good: `active`, `inactive`, `pending`
   - Avoid: `a`, `i`, `p` (too cryptic)

4. **Avoid frequent changes**: Since you may need to maintain mappings (Option 2), try to stabilize your enum values early in development.

## Further Reading

- [Protocol Buffers Enum Documentation](https://protobuf.dev/programming-guides/proto3/#enum)
- [Proto3 JSON Mapping](https://protobuf.dev/programming-guides/proto3/#json)
