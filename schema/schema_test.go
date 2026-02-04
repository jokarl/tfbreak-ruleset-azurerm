package schema

import (
	"sort"
	"testing"
)

func TestSchema_Load(t *testing.T) {
	schema := Load()
	if schema == nil {
		t.Fatal("Load() returned nil")
	}
	if schema.ResourceSchemas == nil {
		t.Fatal("ResourceSchemas is nil")
	}
}

func TestSchema_GetForceNew_ResourceGroup(t *testing.T) {
	schema := Load()

	attrs := schema.GetForceNewAttributes("azurerm_resource_group")
	if len(attrs) == 0 {
		t.Fatal("Expected ForceNew attributes for azurerm_resource_group")
	}

	// Sort for consistent comparison
	sort.Strings(attrs)

	// Check for expected attributes
	expected := []string{"location", "name"}
	sort.Strings(expected)

	if len(attrs) != len(expected) {
		t.Errorf("Expected %d ForceNew attributes, got %d: %v", len(expected), len(attrs), attrs)
	}

	for i, attr := range attrs {
		if i < len(expected) && attr != expected[i] {
			t.Errorf("Expected attribute %s, got %s", expected[i], attr)
		}
	}
}

func TestSchema_GetForceNew_NonExistent(t *testing.T) {
	schema := Load()

	attrs := schema.GetForceNewAttributes("nonexistent_resource")
	if len(attrs) != 0 {
		t.Errorf("Expected no attributes for nonexistent resource, got %v", attrs)
	}
}

func TestSchema_HasResource(t *testing.T) {
	schema := Load()

	if !schema.HasResource("azurerm_resource_group") {
		t.Error("Expected HasResource to return true for azurerm_resource_group")
	}

	if schema.HasResource("nonexistent_resource") {
		t.Error("Expected HasResource to return false for nonexistent resource")
	}
}

func TestSchema_IsForceNew(t *testing.T) {
	schema := Load()

	tests := []struct {
		resourceType string
		attribute    string
		expected     bool
	}{
		{"azurerm_resource_group", "name", true},
		{"azurerm_resource_group", "location", true},
		{"azurerm_resource_group", "tags", false},
		{"azurerm_storage_account", "name", true},
		{"azurerm_storage_account", "account_replication_type", false},
		{"nonexistent_resource", "any", false},
	}

	for _, tt := range tests {
		t.Run(tt.resourceType+"."+tt.attribute, func(t *testing.T) {
			result := schema.IsForceNew(tt.resourceType, tt.attribute)
			if result != tt.expected {
				t.Errorf("IsForceNew(%s, %s) = %v, want %v",
					tt.resourceType, tt.attribute, result, tt.expected)
			}
		})
	}
}

func TestLoadFromJSON(t *testing.T) {
	jsonData := []byte(`{
		"resource_schemas": {
			"test_resource": {
				"block": {
					"attributes": {
						"id": {"type": "string", "computed": true},
						"name": {"type": "string", "required": true, "force_new": true}
					}
				}
			}
		}
	}`)

	schema, err := LoadFromJSON(jsonData)
	if err != nil {
		t.Fatalf("LoadFromJSON failed: %v", err)
	}

	if !schema.HasResource("test_resource") {
		t.Error("Expected test_resource to exist")
	}

	attrs := schema.GetForceNewAttributes("test_resource")
	if len(attrs) != 1 || attrs[0] != "name" {
		t.Errorf("Expected [name], got %v", attrs)
	}
}

func TestLoadFromJSON_Invalid(t *testing.T) {
	_, err := LoadFromJSON([]byte("invalid json"))
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestSchema_GetResourceTypes(t *testing.T) {
	jsonData := []byte(`{
		"resource_schemas": {
			"test_resource_a": {"block": {}},
			"test_resource_b": {"block": {}},
			"test_resource_c": {"block": {}}
		}
	}`)

	schema, err := LoadFromJSON(jsonData)
	if err != nil {
		t.Fatalf("LoadFromJSON failed: %v", err)
	}

	types := schema.GetResourceTypes()
	if len(types) != 3 {
		t.Errorf("Expected 3 resource types, got %d: %v", len(types), types)
	}

	// Convert to map for easier checking
	typeMap := make(map[string]bool)
	for _, rt := range types {
		typeMap[rt] = true
	}

	if !typeMap["test_resource_a"] || !typeMap["test_resource_b"] || !typeMap["test_resource_c"] {
		t.Errorf("Missing expected resource types in %v", types)
	}
}

func TestSchema_GetForceNew_NestedBlocks(t *testing.T) {
	jsonData := []byte(`{
		"resource_schemas": {
			"test_resource": {
				"block": {
					"attributes": {
						"name": {"type": "string", "force_new": true}
					},
					"block_types": {
						"nested_block": {
							"nesting_mode": "list",
							"block": {
								"attributes": {
									"nested_attr": {"type": "string", "force_new": true},
									"other_attr": {"type": "string"}
								}
							}
						}
					}
				}
			}
		}
	}`)

	schema, err := LoadFromJSON(jsonData)
	if err != nil {
		t.Fatalf("LoadFromJSON failed: %v", err)
	}

	attrs := schema.GetForceNewAttributes("test_resource")

	// Should have "name" and "nested_block.nested_attr"
	attrMap := make(map[string]bool)
	for _, a := range attrs {
		attrMap[a] = true
	}

	if !attrMap["name"] {
		t.Error("Expected 'name' in ForceNew attributes")
	}
	if !attrMap["nested_block.nested_attr"] {
		t.Error("Expected 'nested_block.nested_attr' in ForceNew attributes")
	}
	if attrMap["nested_block.other_attr"] {
		t.Error("Did not expect 'nested_block.other_attr' in ForceNew attributes")
	}
}

func TestSchema_IsForceNew_NestedPath(t *testing.T) {
	jsonData := []byte(`{
		"resource_schemas": {
			"test_resource": {
				"block": {
					"block_types": {
						"nested": {
							"nesting_mode": "list",
							"block": {
								"attributes": {
									"force_new_attr": {"type": "string", "force_new": true}
								}
							}
						}
					}
				}
			}
		}
	}`)

	schema, err := LoadFromJSON(jsonData)
	if err != nil {
		t.Fatalf("LoadFromJSON failed: %v", err)
	}

	if !schema.IsForceNew("test_resource", "nested.force_new_attr") {
		t.Error("Expected 'nested.force_new_attr' to be ForceNew")
	}

	// Asking for the block itself should return false
	if schema.IsForceNew("test_resource", "nested") {
		t.Error("Did not expect 'nested' (the block itself) to be ForceNew")
	}
}

func TestSchema_GetForceNew_NilBlock(t *testing.T) {
	jsonData := []byte(`{
		"resource_schemas": {
			"test_resource": {}
		}
	}`)

	schema, err := LoadFromJSON(jsonData)
	if err != nil {
		t.Fatalf("LoadFromJSON failed: %v", err)
	}

	attrs := schema.GetForceNewAttributes("test_resource")
	if len(attrs) != 0 {
		t.Errorf("Expected no attributes for resource with nil block, got %v", attrs)
	}
}

func TestSchema_IsForceNew_NilBlock(t *testing.T) {
	jsonData := []byte(`{
		"resource_schemas": {
			"test_resource": {}
		}
	}`)

	schema, err := LoadFromJSON(jsonData)
	if err != nil {
		t.Fatalf("LoadFromJSON failed: %v", err)
	}

	if schema.IsForceNew("test_resource", "any_attr") {
		t.Error("Expected false for resource with nil block")
	}
}

func TestSchema_IsForceNew_NilNestedBlock(t *testing.T) {
	jsonData := []byte(`{
		"resource_schemas": {
			"test_resource": {
				"block": {
					"block_types": {
						"nested": {
							"nesting_mode": "list"
						}
					}
				}
			}
		}
	}`)

	schema, err := LoadFromJSON(jsonData)
	if err != nil {
		t.Fatalf("LoadFromJSON failed: %v", err)
	}

	if schema.IsForceNew("test_resource", "nested.any_attr") {
		t.Error("Expected false when nested block has nil Block field")
	}
}
