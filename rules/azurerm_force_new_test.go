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
