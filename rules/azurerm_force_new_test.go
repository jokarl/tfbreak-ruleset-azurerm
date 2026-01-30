package rules

import (
	"strings"
	"testing"

	"github.com/jokarl/tfbreak-plugin-sdk/helper"
	"github.com/jokarl/tfbreak-plugin-sdk/tflint"
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
