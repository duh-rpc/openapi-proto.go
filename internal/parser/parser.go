package parser

import (
	"fmt"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// Document wraps the libopenapi v3 document model
type Document struct {
	model *libopenapi.DocumentModel[v3.Document]
}

// SchemaEntry represents a schema with its name and proxy
type SchemaEntry struct {
	Name  string
	Proxy *base.SchemaProxy
}

// ParseDocument parses OpenAPI bytes and returns the document.
// It validates that the document is OpenAPI 3.x and handles both YAML and JSON formats.
func ParseDocument(openapi []byte) (*Document, error) {
	doc, err := libopenapi.NewDocument(openapi)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI document: %w", err)
	}

	model, errs := doc.BuildV3Model()
	if errs != nil {
		return nil, fmt.Errorf("failed to build OpenAPI model: %w", errs)
	}

	if model == nil {
		return nil, fmt.Errorf("only OpenAPI 3.x is supported")
	}

	return &Document{model: model}, nil
}

// Schemas returns schemas from components/schemas in insertion order.
// Returns an empty slice if there are no schemas defined.
func (d *Document) Schemas() ([]*SchemaEntry, error) {
	if d.model.Model.Components == nil {
		return []*SchemaEntry{}, nil
	}

	if d.model.Model.Components.Schemas == nil {
		return []*SchemaEntry{}, nil
	}

	var entries []*SchemaEntry
	for name, proxy := range d.model.Model.Components.Schemas.FromOldest() {
		entries = append(entries, &SchemaEntry{
			Name:  name,
			Proxy: proxy,
		})
	}

	return entries, nil
}
