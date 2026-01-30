// Package schema provides Azure RM provider schema loading and querying.
package schema

import (
	"compress/gzip"
	"embed"
	"encoding/json"
	"io"
	"strings"
	"sync"
)

//go:embed azurerm.json.gz
var embeddedSchema embed.FS

// Schema represents the Azure RM provider schema.
type Schema struct {
	ResourceSchemas map[string]*ResourceSchema `json:"resource_schemas"`
}

// ResourceSchema represents the schema for a single resource type.
type ResourceSchema struct {
	Block *BlockSchema `json:"block"`
}

// BlockSchema represents a block within a resource schema.
type BlockSchema struct {
	Attributes map[string]*AttributeSchema `json:"attributes,omitempty"`
	BlockTypes map[string]*NestedBlockSchema `json:"block_types,omitempty"`
}

// AttributeSchema represents an attribute within a block.
type AttributeSchema struct {
	Type        interface{} `json:"type,omitempty"`
	Description string      `json:"description,omitempty"`
	Required    bool        `json:"required,omitempty"`
	Optional    bool        `json:"optional,omitempty"`
	Computed    bool        `json:"computed,omitempty"`
	ForceNew    bool        `json:"force_new,omitempty"`
	Sensitive   bool        `json:"sensitive,omitempty"`
	Deprecated  string      `json:"deprecated,omitempty"`
}

// NestedBlockSchema represents a nested block within a resource.
type NestedBlockSchema struct {
	NestingMode string       `json:"nesting_mode"`
	Block       *BlockSchema `json:"block"`
	MinItems    int          `json:"min_items,omitempty"`
	MaxItems    int          `json:"max_items,omitempty"`
}

var (
	schemaInstance *Schema
	schemaOnce     sync.Once
	schemaErr      error
)

// Load returns the embedded Azure RM provider schema.
// The schema is loaded lazily and cached for subsequent calls.
func Load() *Schema {
	schemaOnce.Do(func() {
		schemaInstance, schemaErr = loadFromEmbedded()
	})
	if schemaErr != nil {
		// Return an empty schema if loading fails.
		// This allows the plugin to still function, just without ForceNew detection.
		return &Schema{
			ResourceSchemas: make(map[string]*ResourceSchema),
		}
	}
	return schemaInstance
}

// loadFromEmbedded loads the schema from the embedded gzip file.
func loadFromEmbedded() (*Schema, error) {
	f, err := embeddedSchema.Open("azurerm.json.gz")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer gzr.Close()

	data, err := io.ReadAll(gzr)
	if err != nil {
		return nil, err
	}

	var schema Schema
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, err
	}

	return &schema, nil
}

// LoadFromJSON loads the schema from JSON data (for testing).
func LoadFromJSON(data []byte) (*Schema, error) {
	var schema Schema
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, err
	}
	return &schema, nil
}

// GetForceNewAttributes returns the list of ForceNew attribute names for a resource type.
// It searches both top-level attributes and nested blocks.
func (s *Schema) GetForceNewAttributes(resourceType string) []string {
	rs, ok := s.ResourceSchemas[resourceType]
	if !ok {
		return nil
	}
	if rs.Block == nil {
		return nil
	}

	var attrs []string
	attrs = append(attrs, getForceNewFromBlock(rs.Block, "")...)
	return attrs
}

// getForceNewFromBlock recursively finds ForceNew attributes in a block.
func getForceNewFromBlock(block *BlockSchema, prefix string) []string {
	var attrs []string

	// Check attributes
	for name, attr := range block.Attributes {
		if attr.ForceNew {
			fullName := name
			if prefix != "" {
				fullName = prefix + "." + name
			}
			attrs = append(attrs, fullName)
		}
	}

	// Check nested blocks
	for name, nested := range block.BlockTypes {
		if nested.Block != nil {
			nestedPrefix := name
			if prefix != "" {
				nestedPrefix = prefix + "." + name
			}
			attrs = append(attrs, getForceNewFromBlock(nested.Block, nestedPrefix)...)
		}
	}

	return attrs
}

// HasResource checks if a resource type exists in the schema.
func (s *Schema) HasResource(resourceType string) bool {
	_, ok := s.ResourceSchemas[resourceType]
	return ok
}

// GetResourceTypes returns all resource types in the schema.
func (s *Schema) GetResourceTypes() []string {
	types := make([]string, 0, len(s.ResourceSchemas))
	for t := range s.ResourceSchemas {
		types = append(types, t)
	}
	return types
}

// IsForceNew checks if a specific attribute is ForceNew for a resource type.
func (s *Schema) IsForceNew(resourceType, attributePath string) bool {
	rs, ok := s.ResourceSchemas[resourceType]
	if !ok {
		return false
	}
	if rs.Block == nil {
		return false
	}

	return isForceNewInBlock(rs.Block, attributePath)
}

// isForceNewInBlock checks if an attribute path is ForceNew in a block.
func isForceNewInBlock(block *BlockSchema, path string) bool {
	parts := strings.SplitN(path, ".", 2)
	name := parts[0]

	// Check if it's a direct attribute
	if attr, ok := block.Attributes[name]; ok {
		if len(parts) == 1 {
			return attr.ForceNew
		}
		// Can't descend into attribute
		return false
	}

	// Check if it's a nested block
	if nested, ok := block.BlockTypes[name]; ok {
		if nested.Block == nil {
			return false
		}
		if len(parts) == 1 {
			// The block itself - check if any attribute in it is ForceNew
			return false
		}
		return isForceNewInBlock(nested.Block, parts[1])
	}

	return false
}
