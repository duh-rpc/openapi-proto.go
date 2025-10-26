package conv_test

import (
	"testing"

	conv "github.com/duh-rpc/openapi-proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertBasics(t *testing.T) {
	for _, test := range []struct {
		name     string
		given    []byte
		pkg      string
		expected string
		wantErr  string
	}{
		{
			name:    "empty openapi bytes",
			given:   []byte{},
			pkg:     "testpkg",
			wantErr: "openapi input cannot be empty",
		},
		{
			name:    "empty package name",
			given:   []byte("openapi: 3.0.0"),
			pkg:     "",
			wantErr: "package name cannot be empty",
		},
		{
			name:    "invalid YAML syntax",
			given:   []byte("this is not valid: [yaml"),
			pkg:     "testpkg",
			wantErr: "failed to parse OpenAPI document",
		},
		{
			name: "valid minimal OpenAPI 3.0",
			given: []byte(`openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
`),
			pkg:      "testpkg",
			expected: "syntax = \"proto3\";\n\npackage testpkg;\n",
		},
		{
			name: "OpenAPI 2.0 Swagger",
			given: []byte(`swagger: "2.0"
info:
  title: Test API
  version: 1.0.0
paths: {}
`),
			pkg:     "testpkg",
			wantErr: "supplied spec is a different version",
		},
		{
			name: "valid JSON OpenAPI",
			given: []byte(`{
  "openapi": "3.0.0",
  "info": {
    "title": "Test API",
    "version": "1.0.0"
  },
  "paths": {}
}`),
			pkg:      "testpkg",
			expected: "syntax = \"proto3\";\n\npackage testpkg;\n",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert(test.given, test.pkg)

			if test.wantErr != "" {
				require.ErrorContains(t, err, test.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expected, string(result))
		})
	}
}

func TestConvertParseDocument(t *testing.T) {
	for _, test := range []struct {
		name     string
		given    string
		pkg      string
		expected string
		wantErr  string
	}{
		{
			name: "parse valid OpenAPI 3.0 YAML",
			given: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
`,
			pkg:      "testpkg",
			expected: "syntax = \"proto3\";\n\npackage testpkg;\n",
		},
		{
			name: "parse valid OpenAPI 3.0 JSON",
			given: `{
  "openapi": "3.0.0",
  "info": {
    "title": "Test API",
    "version": "1.0.0"
  },
  "paths": {}
}`,
			pkg:      "testpkg",
			expected: "syntax = \"proto3\";\n\npackage testpkg;\n",
		},
		{
			name:    "non-OpenAPI document",
			given:   `title: Some Random YAML`,
			pkg:     "testpkg",
			wantErr: "spec type not supported",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.given), test.pkg)

			if test.wantErr != "" {
				require.ErrorContains(t, err, test.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expected, string(result))
		})
	}
}

func TestConvertExtractSchemas(t *testing.T) {
	for _, test := range []struct {
		name     string
		given    string
		pkg      string
		expected string
	}{
		{
			name: "extract schemas in YAML insertion order",
			given: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas:
    User:
      type: object
    Product:
      type: object
    Order:
      type: object
`,
			pkg:      "testpkg",
			expected: "syntax = \"proto3\";\n\npackage testpkg;\n",
		},
		{
			name: "document with no components section",
			given: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
`,
			pkg:      "testpkg",
			expected: "syntax = \"proto3\";\n\npackage testpkg;\n",
		},
		{
			name: "document with empty components/schemas",
			given: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas: {}
`,
			pkg:      "testpkg",
			expected: "syntax = \"proto3\";\n\npackage testpkg;\n",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.given), test.pkg)

			require.NoError(t, err)
			assert.Equal(t, test.expected, string(result))
		})
	}
}
