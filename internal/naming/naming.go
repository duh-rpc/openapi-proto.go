package naming

import (
	"fmt"
	"strings"
	"unicode"
)

// ToSnakeCase converts camelCase/PascalCase to snake_case.
// Algorithm: Each uppercase letter becomes lowercase with underscore prefix (except first char).
// Examples: userId → user_id, HTTPStatus → h_t_t_p_status, email → email
func ToSnakeCase(s string) string {
	if s == "" {
		return ""
	}

	var result strings.Builder
	result.Grow(len(s) + 5)

	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteRune('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// ToPascalCase converts snake_case/camelCase to PascalCase.
// Examples: user_id → UserId, shippingAddress → ShippingAddress
func ToPascalCase(s string) string {
	if s == "" {
		return ""
	}

	var result strings.Builder
	result.Grow(len(s))

	capitalizeNext := true
	for _, r := range s {
		if r == '_' {
			capitalizeNext = true
			continue
		}

		if capitalizeNext {
			result.WriteRune(unicode.ToUpper(r))
			capitalizeNext = false
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// ToEnumValueName converts a value to ENUM_PREFIX_VALUE_NAME format.
// Examples: (Status, active) → STATUS_ACTIVE, (Status, in-progress) → STATUS_IN_PROGRESS, (Code, 401) → CODE_401
func ToEnumValueName(enumName, value string) string {
	upperEnum := strings.ToUpper(ToSnakeCase(enumName))
	upperValue := strings.ToUpper(value)
	upperValue = strings.ReplaceAll(upperValue, "-", "_")
	return fmt.Sprintf("%s_%s", upperEnum, upperValue)
}

// NeedsJSONName returns true if proto field name differs from original.
func NeedsJSONName(original, protoField string) bool {
	return ToSnakeCase(original) != original
}

// NameTracker tracks used names and generates unique names when conflicts occur.
type NameTracker struct {
	used map[string]int
}

// NewNameTracker creates a new NameTracker.
func NewNameTracker() *NameTracker {
	return &NameTracker{
		used: make(map[string]int),
	}
}

// UniqueName returns a unique name, adding numeric suffix if needed (_2, _3, etc.).
func (nt *NameTracker) UniqueName(name string) string {
	count, exists := nt.used[name]
	if !exists {
		nt.used[name] = 1
		return name
	}

	count++
	nt.used[name] = count
	return fmt.Sprintf("%s_%d", name, count)
}
