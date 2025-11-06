# Discriminated Unions and OneOf

This document explains why OpenAPI `oneOf` with discriminators cannot be converted to Protocol Buffer `oneof` fields, and what alternatives you can use instead.

## Overview

**Discriminated unions are not supported** by this library. OpenAPI's `oneOf` with discriminators and Protocol Buffer's `oneof` produce **fundamentally incompatible JSON formats**. A client generated from an OpenAPI specification cannot communicate with a server using the protobuf definition due to structural differences in the JSON representation.

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

## Alternative Approaches

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

1. **Design for compatibility**: If you need both OpenAPI and protobuf definitions, avoid `oneOf` patterns in your API design from the start.

2. **Use Option 2 for polymorphism**: The flexible schema approach (Option 2) provides the best balance of compatibility and functionality.

3. **Document your choice**: Whichever approach you choose, document it clearly in your API documentation so clients know what to expect.

4. **Consider your use case**: If you only need one format (either REST with OpenAPI OR RPC with protobuf), you don't need to worry about this limitation.

## Related Issues

This limitation affects any schema composition features that would map to protobuf `oneof`:

- ❌ `oneOf` with discriminators
- ❌ `anyOf` (would also map to `oneof`)
- ❌ Polymorphic schemas with inheritance

For now, these features are not supported by this library. If you have a compelling use case that requires these features, please open an issue to discuss potential solutions.

## Further Reading

- [Protocol Buffers JSON Mapping](https://protobuf.dev/programming-guides/json/)
- [OpenAPI Discriminator Specification](https://spec.openapis.org/oas/v3.0.3#discriminator-object)
- [Stack Overflow: Protobuf oneof JSON Serialization](https://stackoverflow.com/questions/44069789/what-is-the-expected-json-serialization-of-oneof-protobuf-field)
