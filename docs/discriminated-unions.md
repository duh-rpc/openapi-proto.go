# Discriminated Unions and OneOf

This document explains how OpenAPI `oneOf` with discriminators is handled by this library, why it cannot be converted to Protocol Buffer `oneof` fields, and the Go code generation approach used instead.

## Overview

**OneOf with discriminators is now supported** through Go code generation. OpenAPI's `oneOf` with discriminators and Protocol Buffer's `oneof` produce **fundamentally incompatible JSON formats**. To maintain complete JSON compatibility with the OpenAPI specification, schemas containing unions are generated as **Go structs with custom JSON marshaling** instead of protobuf messages.

This approach:
- ✅ Maintains exact OpenAPI JSON format compatibility
- ✅ Provides type-safe union handling in Go
- ✅ Supports case-insensitive discriminator matching
- ✅ Automatically handles transitive dependencies (types referencing unions)

## The Incompatibility

### OpenAPI oneOf JSON Format

When you define a union with a discriminator in OpenAPI, the JSON format is **flat** - the discriminator is just another property within the object:

**OpenAPI Schema:**
```yaml
components:
  schemas:
    Pet:
      oneOf:
        - $ref: '#/components/schemas/Dog'
        - $ref: '#/components/schemas/Cat'
      discriminator:
        propertyName: petType

    Dog:
      type: object
      required: [petType, bark]
      properties:
        petType:
          type: string
          enum: [dog]
        bark:
          type: string

    Cat:
      type: object
      required: [petType, meow]
      properties:
        petType:
          type: string
          enum: [cat]
        meow:
          type: string
```

**JSON Representation:**
```json
{
  "petType": "dog",
  "bark": "woof"
}
```

The discriminator field (`petType`) appears at the same level as other properties. This is how OpenAPI-generated clients (like oapi-codegen, OpenAPI Generator, etc.) serialize union types.

### Protobuf oneof JSON Format

When you define a `oneof` in protobuf, the JSON format includes a **wrapper** - the selected variant is nested under the oneof field name:

**Protobuf Schema:**
```protobuf
message Pet {
  oneof pet_type {
    Dog dog = 1;
    Cat cat = 2;
  }
}

message Dog {
  string pet_type = 1 [json_name = "petType"];
  string bark = 2 [json_name = "bark"];
}

message Cat {
  string pet_type = 1 [json_name = "petType"];
  string meow = 2 [json_name = "meow"];
}
```

**JSON Representation (using protojson):**
```json
{
  "dog": {
    "petType": "dog",
    "bark": "woof"
  }
}
```

The outer `"dog"` key indicates which oneof field is set, and the actual object data is nested inside. This wrapper is required by the [Protocol Buffers JSON specification](https://protobuf.dev/programming-guides/json/).

### Why This Breaks Client-Server Communication

When you try to use them together:

1. **OpenAPI-generated client** sends:
   ```json
   {"petType": "dog", "bark": "woof"}
   ```

2. **Protobuf-based server** expects:
   ```json
   {"dog": {"petType": "dog", "bark": "woof"}}
   ```

3. **Result**: Deserialization fails ❌

The protobuf unmarshaler cannot parse the flat JSON format because it expects the wrapper object. Similarly, an OpenAPI client cannot parse the wrapped format the protobuf server returns.

## Go Code Generation Solution

This library solves the incompatibility by generating Go structs with custom JSON marshaling for union types. The approach follows the **oapi-codegen pattern** where union types contain pointer fields to each variant, with custom `MarshalJSON` and `UnmarshalJSON` methods that produce the flat OpenAPI JSON format.

### Generated Go Code Pattern

For a `Pet` union with `Dog` and `Cat` variants:

**OpenAPI Schema:**
```yaml
Pet:
  oneOf:
    - $ref: '#/components/schemas/Dog'
    - $ref: '#/components/schemas/Cat'
  discriminator:
    propertyName: petType

Dog:
  type: object
  properties:
    petType:
      type: string
      enum: [dog]
    bark:
      type: string

Cat:
  type: object
  properties:
    petType:
      type: string
      enum: [cat]
    meow:
      type: string
```

**Generated Go Code:**
```go
package mypkg

import (
    "encoding/json"
    "fmt"
    "strings"
)

// Union wrapper with pointer fields to variants
type Pet struct {
    Dog *Dog
    Cat *Cat
}

// Custom marshaling to match flat OpenAPI JSON
func (u *Pet) MarshalJSON() ([]byte, error) {
    if u.Dog != nil {
        return json.Marshal(u.Dog)
    }
    if u.Cat != nil {
        return json.Marshal(u.Cat)
    }
    return nil, fmt.Errorf("Pet: no variant set")
}

func (u *Pet) UnmarshalJSON(data []byte) error {
    var discriminator struct {
        PetType string `json:"petType"`
    }
    if err := json.Unmarshal(data, &discriminator); err != nil {
        return err
    }

    // Case-insensitive discriminator matching
    switch strings.ToLower(discriminator.PetType) {
    case "dog":
        u.Dog = &Dog{}
        return json.Unmarshal(data, u.Dog)
    case "cat":
        u.Cat = &Cat{}
        return json.Unmarshal(data, u.Cat)
    default:
        return fmt.Errorf("unknown petType: %s", discriminator.PetType)
    }
}

// Variant types as separate structs
type Dog struct {
    PetType string `json:"petType"`
    Bark    string `json:"bark"`
}

type Cat struct {
    PetType string `json:"petType"`
    Meow    string `json:"meow"`
}
```

**JSON Format** (matches OpenAPI spec exactly):
```json
{"petType": "dog", "bark": "woof"}
```

### Transitive Closure Behavior

When a schema contains or references a union, it must be generated as Go code (not protobuf). This applies transitively:

**Example Schema Set:**
```yaml
Address:
  type: object
  properties:
    street: {type: string}

Pet:
  oneOf:
    - $ref: '#/components/schemas/Dog'
    - $ref: '#/components/schemas/Cat'
  discriminator:
    propertyName: petType

Dog:
  type: object
  properties:
    petType: {type: string}
    bark: {type: string}

Cat:
  type: object
  properties:
    petType: {type: string}
    meow: {type: string}

Owner:
  type: object
  properties:
    name: {type: string}
    pet:
      $ref: '#/components/schemas/Pet'
```

**Classification Result:**
- **Address** → Proto (no union connection)
- **Pet** → Go (contains oneOf)
- **Dog** → Go (variant of union type Pet)
- **Cat** → Go (variant of union type Pet)
- **Owner** → Go (references union type Pet)

The `ConvertResult.TypeMap` provides complete visibility:
```go
result, err := conv.Convert(openapi, opts)

for typeName, info := range result.TypeMap {
    fmt.Printf("%s: %s (%s)\n", typeName, info.Location, info.Reason)
}
// Output:
// Address: proto ()
// Pet: golang (contains oneOf)
// Dog: golang (variant of union type Pet)
// Cat: golang (variant of union type Pet)
// Owner: golang (references union type Pet)
```

### What's Supported

**Phase 1 Support (Current):**
- ✅ `oneOf` with discriminators
- ✅ `$ref`-based variants only (no inline schemas)
- ✅ Case-insensitive discriminator matching
- ✅ Explicit discriminator.mapping overrides
- ✅ Transitive closure (types referencing unions become Go types)
- ✅ Custom JSON marshaling/unmarshaling

**Not Supported (Future Phases):**
- ❌ `oneOf` without discriminators
- ❌ Inline oneOf variants (must use `$ref`)
- ❌ `anyOf`, `allOf`, `not`
- ❌ Nested unions (unions within unions)

### Requirements

For a oneOf schema to be accepted:

1. **Must have discriminator**: `discriminator.propertyName` is required
2. **Variants must use $ref**: All oneOf items must be `$ref` references
3. **Discriminator in variants**: Each variant must include the discriminator property
4. **Case-insensitive matching**: Discriminator values are matched case-insensitively to schema names

**Valid Example:**
```yaml
Pet:
  oneOf:
    - $ref: '#/components/schemas/Dog'
    - $ref: '#/components/schemas/Cat'
  discriminator:
    propertyName: petType

Dog:
  type: object
  properties:
    petType:
      type: string
      enum: [dog]
    bark:
      type: string
```

**Invalid Examples:**
```yaml
# Missing discriminator - ERROR
Pet:
  oneOf:
    - $ref: '#/components/schemas/Dog'
    - $ref: '#/components/schemas/Cat'

# Inline variant - ERROR
Pet:
  oneOf:
    - type: object
      properties:
        bark: {type: string}
  discriminator:
    propertyName: petType
```

## Alternative Approaches (If Go Generation Doesn't Work)

Since discriminated unions cannot be directly converted, here are alternative patterns you can use:

### Option 1: Avoid oneOf in Your API Design

The simplest solution is to not use `oneOf` in your OpenAPI specification. Consider alternative designs:

**Instead of this:**
```yaml
Pet:
  oneOf:
    - $ref: '#/components/schemas/Dog'
    - $ref: '#/components/schemas/Cat'
  discriminator:
    propertyName: petType
```

**Use separate endpoints:**
```yaml
paths:
  /pets/dogs:
    post:
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Dog'
  /pets/cats:
    post:
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Cat'
```

**Pros**: Complete compatibility, clear API semantics.

**Cons**: More endpoints to maintain.

### Option 2: Use a Single Flexible Schema

Instead of multiple schemas with discriminators, use a single schema with optional fields:

```yaml
Pet:
  type: object
  required: [petType]
  properties:
    petType:
      type: string
      enum: [dog, cat]
    bark:
      type: string
      description: "Only present for dogs"
    meow:
      type: string
      description: "Only present for cats"
```

**Protobuf:**
```protobuf
message Pet {
  string pet_type = 1 [json_name = "petType"];
  optional string bark = 2 [json_name = "bark"];
  optional string meow = 3 [json_name = "meow"];
}
```

**JSON:**
```json
{
  "petType": "dog",
  "bark": "woof"
}
```

**Pros**: Perfectly compatible JSON format, simple implementation.

**Cons**: No compile-time type safety, unused fields present in the schema.

### Option 3: Separate OpenAPI and Protobuf Specs

Accept that the two formats serve different purposes and maintain separate source specifications:

- **OpenAPI spec**: For REST API documentation and client generation, use `oneOf` with discriminators as needed
- **Protobuf spec**: For RPC definitions and strongly-typed serialization, use `oneof` or separate messages

**Pros**: Use each format's strengths without compromise.

**Cons**: Two specs to maintain, requires careful synchronization.

### Option 4: Custom JSON Marshaling (Advanced)

Implement custom JSON marshaling/unmarshaling in your protobuf code that transforms the flat OpenAPI format to/from the wrapped protobuf format. This requires writing custom code in your application.

**Example (Go):**
```go
func (p *Pet) UnmarshalJSON(data []byte) error {
    // Parse flat OpenAPI format
    var flat struct {
        PetType string `json:"petType"`
        Bark    string `json:"bark"`
        Meow    string `json:"meow"`
    }

    if err := json.Unmarshal(data, &flat); err != nil {
        return err
    }

    // Transform to protobuf oneof structure
    switch flat.PetType {
    case "dog":
        p.PetType = &Pet_Dog{Dog: &Dog{
            PetType: flat.PetType,
            Bark:    flat.Bark,
        }}
    case "cat":
        p.PetType = &Pet_Cat{Cat: &Cat{
            PetType: flat.PetType,
            Meow:    flat.Meow,
        }}
    }

    return nil
}
```

**Pros**: Maintains both the OpenAPI contract and protobuf type safety.

**Cons**: Complex, requires custom code for each union type, error-prone.

## Why Not Auto-Generate the Conversion?

You might wonder why this library doesn't automatically handle the conversion. Here are the reasons:

1. **Semantic Mismatch**: The protobuf wrapper format fundamentally changes the JSON structure. Auto-converting would create a protobuf file that doesn't produce the JSON format your OpenAPI spec defines.

2. **Breaking Contract**: The OpenAPI specification defines an exact JSON contract. Generating protobuf with `oneof` would break that contract, causing integration failures.

3. **Hidden Complexity**: Auto-conversion would hide the incompatibility from users, leading to runtime failures that are hard to debug.

4. **Better Alternatives**: Options 1 and 2 (separate endpoints or flexible schemas) are simpler and more maintainable than complex marshaling code.

## Best Practices

1. **Use Go generation for unions**: The library's Go code generation with custom marshaling maintains perfect OpenAPI JSON compatibility while providing type safety.

2. **Check TypeMap**: Use `ConvertResult.TypeMap` to understand which types are generated as Go vs protobuf and why.

3. **Meet requirements**: Ensure your oneOf schemas have discriminators and use `$ref`-based variants (not inline schemas).

4. **Consider alternatives**: If Go generation doesn't fit your use case, the alternative approaches below may work better for your specific scenario.

## Related Schema Composition Features

Schema composition support status:

- ✅ `oneOf` with discriminators (generates Go code)
- ❌ `oneOf` without discriminators (requires discriminator)
- ❌ `anyOf` (future phase)
- ❌ `allOf` (future phase)
- ❌ `not` (future phase)

If you have a compelling use case that requires unsupported features, please open an issue to discuss potential solutions.

## Further Reading

- [Protocol Buffers JSON Mapping](https://protobuf.dev/programming-guides/json/)
- [OpenAPI Discriminator Specification](https://spec.openapis.org/oas/v3.0.3#discriminator-object)
- [Stack Overflow: Protobuf oneof JSON Serialization](https://stackoverflow.com/questions/44069789/what-is-the-expected-json-serialization-of-oneof-protobuf-field)
