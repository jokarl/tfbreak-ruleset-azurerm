package rules

import (
	"strings"
	"testing"

	"github.com/jokarl/tfbreak-plugin-sdk/helper"
	"github.com/jokarl/tfbreak-plugin-sdk/hclext"
	"github.com/jokarl/tfbreak-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
)

func TestForceNew_Name(t *testing.T) {
	rule := NewAzurermForceNewRule()
	if rule.Name() != "azurerm_force_new" {
		t.Errorf("Expected rule name to be 'azurerm_force_new', got '%s'", rule.Name())
	}
}

func TestForceNew_Enabled(t *testing.T) {
	rule := NewAzurermForceNewRule()
	if !rule.Enabled() {
		t.Error("Expected rule to be enabled by default")
	}
}

func TestForceNew_Severity(t *testing.T) {
	rule := NewAzurermForceNewRule()
	if rule.Severity() != tflint.ERROR {
		t.Errorf("Expected severity to be ERROR, got %v", rule.Severity())
	}
}

func TestForceNew_Link(t *testing.T) {
	rule := NewAzurermForceNewRule()
	link := rule.Link()
	if link == "" {
		t.Error("Expected non-empty link")
	}
	// Should contain the rule name
	if !strings.Contains(link, "azurerm_force_new") {
		t.Errorf("Expected link to contain rule name, got '%s'", link)
	}
}

func TestForceNew_DetectsChange(t *testing.T) {
	rule := NewAzurermForceNewRule()

	runner := helper.TestRunner(t,
		map[string]string{
			"main.tf": `
resource "azurerm_resource_group" "example" {
    name     = "my-rg"
    location = "westeurope"
    tags     = {
        env = "dev"
    }
}`,
		},
		map[string]string{
			"main.tf": `
resource "azurerm_resource_group" "example" {
    name     = "my-rg"
    location = "eastus"
    tags     = {
        env = "dev"
    }
}`,
		},
	)

	err := rule.Check(runner)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}

	if len(runner.Issues) == 0 {
		t.Error("Expected issue to be emitted for ForceNew change")
	}

	// Verify the issue contains the attribute name
	if len(runner.Issues) > 0 && !strings.Contains(runner.Issues[0].Message, "location") {
		t.Errorf("Expected issue message to mention 'location', got '%s'", runner.Issues[0].Message)
	}
}

func TestForceNew_NoChange(t *testing.T) {
	rule := NewAzurermForceNewRule()

	runner := helper.TestRunner(t,
		map[string]string{
			"main.tf": `
resource "azurerm_resource_group" "example" {
    name     = "my-rg"
    location = "westeurope"
}`,
		},
		map[string]string{
			"main.tf": `
resource "azurerm_resource_group" "example" {
    name     = "my-rg"
    location = "westeurope"
}`,
		},
	)

	err := rule.Check(runner)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}

	helper.AssertNoIssues(t, runner.Issues)
}

func TestForceNew_NonForceNew(t *testing.T) {
	rule := NewAzurermForceNewRule()

	runner := helper.TestRunner(t,
		map[string]string{
			"main.tf": `
resource "azurerm_resource_group" "example" {
    name     = "my-rg"
    location = "westeurope"
    tags     = {
        env = "dev"
    }
}`,
		},
		map[string]string{
			"main.tf": `
resource "azurerm_resource_group" "example" {
    name     = "my-rg"
    location = "westeurope"
    tags     = {
        env = "prod"
    }
}`,
		},
	)

	err := rule.Check(runner)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}

	// Tags is not a ForceNew attribute, so no issues expected
	helper.AssertNoIssues(t, runner.Issues)
}

func TestForceNew_NewResource(t *testing.T) {
	rule := NewAzurermForceNewRule()

	runner := helper.TestRunner(t,
		map[string]string{
			"main.tf": ``,
		},
		map[string]string{
			"main.tf": `
resource "azurerm_resource_group" "example" {
    name     = "my-rg"
    location = "westeurope"
}`,
		},
	)

	err := rule.Check(runner)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}

	// New resources should not trigger ForceNew warnings
	helper.AssertNoIssues(t, runner.Issues)
}

func TestForceNew_NonAzurermResource(t *testing.T) {
	rule := NewAzurermForceNewRule()

	runner := helper.TestRunner(t,
		map[string]string{
			"main.tf": `
resource "aws_s3_bucket" "example" {
    bucket = "my-bucket"
}`,
		},
		map[string]string{
			"main.tf": `
resource "aws_s3_bucket" "example" {
    bucket = "different-bucket"
}`,
		},
	)

	err := rule.Check(runner)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}

	// Non-azurerm resources should not be checked
	helper.AssertNoIssues(t, runner.Issues)
}

func TestForceNew_AttributeRemoved(t *testing.T) {
	rule := NewAzurermForceNewRule()

	runner := helper.TestRunner(t,
		map[string]string{
			"main.tf": `
resource "azurerm_resource_group" "example" {
    name     = "my-rg"
    location = "westeurope"
}`,
		},
		map[string]string{
			"main.tf": `
resource "azurerm_resource_group" "example" {
    name = "my-rg"
}`,
		},
	)

	err := rule.Check(runner)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}

	// Removing a ForceNew attribute should be detected
	if len(runner.Issues) == 0 {
		t.Error("Expected issue to be emitted when ForceNew attribute is removed")
	}
}

func TestForceNew_AttributeAdded(t *testing.T) {
	rule := NewAzurermForceNewRule()

	runner := helper.TestRunner(t,
		map[string]string{
			"main.tf": `
resource "azurerm_resource_group" "example" {
    name = "my-rg"
}`,
		},
		map[string]string{
			"main.tf": `
resource "azurerm_resource_group" "example" {
    name     = "my-rg"
    location = "westeurope"
}`,
		},
	)

	err := rule.Check(runner)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}

	// Adding a ForceNew attribute should be detected
	if len(runner.Issues) == 0 {
		t.Error("Expected issue to be emitted when ForceNew attribute is added")
	}
}

// =============================================================================
// Unit tests for buildBodySchema and getAttributeByPath
// =============================================================================

func TestBuildBodySchema(t *testing.T) {
	tests := []struct {
		name      string
		paths     []string
		wantAttrs []string
		wantBlocks map[string][]string // block name -> expected nested attrs
	}{
		{
			name:       "top-level only",
			paths:      []string{"location", "name"},
			wantAttrs:  []string{"location", "name"},
			wantBlocks: nil,
		},
		{
			name:       "nested only",
			paths:      []string{"identity.type", "identity.identity_ids"},
			wantAttrs:  nil,
			wantBlocks: map[string][]string{"identity": {"identity_ids", "type"}},
		},
		{
			name:       "mixed top-level and nested",
			paths:      []string{"location", "identity.type", "name"},
			wantAttrs:  []string{"location", "name"},
			wantBlocks: map[string][]string{"identity": {"type"}},
		},
		{
			name:       "deeply nested",
			paths:      []string{"identity.nested.deep_attr"},
			wantAttrs:  nil,
			wantBlocks: map[string][]string{"identity": nil}, // nested block check
		},
		{
			name:       "empty",
			paths:      []string{},
			wantAttrs:  nil,
			wantBlocks: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := buildBodySchema(tt.paths)

			// Check attributes
			gotAttrs := make([]string, len(schema.Attributes))
			for i, attr := range schema.Attributes {
				gotAttrs[i] = attr.Name
			}
			if len(gotAttrs) != len(tt.wantAttrs) {
				t.Errorf("got %d attrs, want %d: %v vs %v", len(gotAttrs), len(tt.wantAttrs), gotAttrs, tt.wantAttrs)
			} else {
				for i, want := range tt.wantAttrs {
					if gotAttrs[i] != want {
						t.Errorf("attr[%d] = %q, want %q", i, gotAttrs[i], want)
					}
				}
			}

			// Check blocks
			if tt.wantBlocks == nil && len(schema.Blocks) > 0 {
				t.Errorf("expected no blocks, got %d", len(schema.Blocks))
			}
			for blockName, wantNestedAttrs := range tt.wantBlocks {
				found := false
				for _, block := range schema.Blocks {
					if block.Type == blockName {
						found = true
						if wantNestedAttrs != nil && block.Body != nil {
							gotNestedAttrs := make([]string, len(block.Body.Attributes))
							for i, attr := range block.Body.Attributes {
								gotNestedAttrs[i] = attr.Name
							}
							if len(gotNestedAttrs) != len(wantNestedAttrs) {
								t.Errorf("block %q: got %d nested attrs, want %d", blockName, len(gotNestedAttrs), len(wantNestedAttrs))
							}
						}
						break
					}
				}
				if !found {
					t.Errorf("expected block %q not found", blockName)
				}
			}
		})
	}
}

func TestGetAttributeByPath(t *testing.T) {
	// Build a test block structure:
	// block {
	//   location = "westeurope"
	//   identity {
	//     type = "SystemAssigned"
	//   }
	// }
	testBlock := &hclext.Block{
		Type:   "azurerm_resource_group",
		Labels: []string{"azurerm_resource_group", "example"},
		Body: &hclext.BodyContent{
			Attributes: map[string]*hclext.Attribute{
				"location": {Name: "location", Value: cty.StringVal("westeurope")},
			},
			Blocks: []*hclext.Block{
				{
					Type: "identity",
					Body: &hclext.BodyContent{
						Attributes: map[string]*hclext.Attribute{
							"type": {Name: "type", Value: cty.StringVal("SystemAssigned")},
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name     string
		block    *hclext.Block
		path     string
		wantName string
		wantNil  bool
	}{
		{
			name:     "top-level attribute",
			block:    testBlock,
			path:     "location",
			wantName: "location",
		},
		{
			name:     "nested attribute",
			block:    testBlock,
			path:     "identity.type",
			wantName: "type",
		},
		{
			name:    "missing top-level attribute",
			block:   testBlock,
			path:    "nonexistent",
			wantNil: true,
		},
		{
			name:    "missing nested block",
			block:   testBlock,
			path:    "nonexistent.attr",
			wantNil: true,
		},
		{
			name:    "nil block",
			block:   nil,
			path:    "location",
			wantNil: true,
		},
		{
			name:    "block with nil body",
			block:   &hclext.Block{Type: "test"},
			path:    "location",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attr := getAttributeByPath(tt.block, tt.path)
			if tt.wantNil {
				if attr != nil {
					t.Errorf("expected nil, got attribute %q", attr.Name)
				}
			} else {
				if attr == nil {
					t.Errorf("expected attribute %q, got nil", tt.wantName)
				} else if attr.Name != tt.wantName {
					t.Errorf("got attribute %q, want %q", attr.Name, tt.wantName)
				}
			}
		})
	}
}

// =============================================================================
// Unit tests for helper functions
// =============================================================================

func TestFormatCtyValue(t *testing.T) {
	tests := []struct {
		name     string
		value    cty.Value
		expected string
	}{
		{"string", cty.StringVal("hello"), "hello"},
		{"number int", cty.NumberIntVal(42), "42"},
		{"number float", cty.NumberFloatVal(3.14), "3.14"},
		{"bool true", cty.BoolVal(true), "true"},
		{"bool false", cty.BoolVal(false), "false"},
		{"null", cty.NullVal(cty.String), "<null>"},
		{"unknown", cty.UnknownVal(cty.String), "<unknown>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatCtyValue(tt.value)
			if result != tt.expected {
				t.Errorf("formatCtyValue(%v) = %q, want %q", tt.value, result, tt.expected)
			}
		})
	}
}

func TestEvalAttr(t *testing.T) {
	t.Run("nil attribute", func(t *testing.T) {
		result := evalAttr(nil)
		if result != "<not set>" {
			t.Errorf("evalAttr(nil) = %q, want %q", result, "<not set>")
		}
	})

	t.Run("attribute with Value", func(t *testing.T) {
		attr := &hclext.Attribute{
			Name:  "test",
			Value: cty.StringVal("from-value"),
		}
		result := evalAttr(attr)
		if result != "from-value" {
			t.Errorf("evalAttr with Value = %q, want %q", result, "from-value")
		}
	})

	t.Run("attribute with null Value", func(t *testing.T) {
		attr := &hclext.Attribute{
			Name:  "test",
			Value: cty.NullVal(cty.String),
		}
		result := evalAttr(attr)
		if result != "<null>" {
			t.Errorf("evalAttr with null Value = %q, want %q", result, "<null>")
		}
	})

	t.Run("attribute with unknown Value", func(t *testing.T) {
		attr := &hclext.Attribute{
			Name:  "test",
			Value: cty.UnknownVal(cty.String),
		}
		result := evalAttr(attr)
		if result != "<unknown>" {
			t.Errorf("evalAttr with unknown Value = %q, want %q", result, "<unknown>")
		}
	})

	t.Run("attribute with no Value or Expr", func(t *testing.T) {
		attr := &hclext.Attribute{
			Name: "test",
			// No Value and no Expr
		}
		result := evalAttr(attr)
		if result != "<dynamic>" {
			t.Errorf("evalAttr with no Value/Expr = %q, want %q", result, "<dynamic>")
		}
	})
}
