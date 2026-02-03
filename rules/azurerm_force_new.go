// Package rules provides the tfbreak rules for Azure RM provider.
package rules

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/jokarl/tfbreak-plugin-sdk/hclext"
	"github.com/jokarl/tfbreak-plugin-sdk/tflint"
	"github.com/jokarl/tfbreak-ruleset-azurerm/project"
	"github.com/jokarl/tfbreak-ruleset-azurerm/schema"
	"github.com/zclconf/go-cty/cty"
)

// AzurermForceNewRule detects changes to ForceNew attributes in Azure RM resources.
// When a ForceNew attribute changes, Terraform will destroy and recreate the resource.
type AzurermForceNewRule struct {
	tflint.DefaultRule
	schema *schema.Schema
}

// NewAzurermForceNewRule creates a new ForceNew detection rule.
func NewAzurermForceNewRule() *AzurermForceNewRule {
	return &AzurermForceNewRule{
		schema: schema.Load(),
	}
}

// Name returns the rule name.
func (r *AzurermForceNewRule) Name() string {
	return "azurerm_force_new"
}

// Enabled returns whether the rule is enabled by default.
func (r *AzurermForceNewRule) Enabled() bool {
	return true
}

// Severity returns the rule severity.
// ForceNew changes are errors because they cause resource destruction.
func (r *AzurermForceNewRule) Severity() tflint.Severity {
	return tflint.ERROR
}

// Link returns the documentation link for this rule.
func (r *AzurermForceNewRule) Link() string {
	return project.ReferenceLink(r.Name())
}

// Check checks for ForceNew attribute changes between old and new configurations.
func (r *AzurermForceNewRule) Check(runner tflint.Runner) error {
	// Get list of azurerm resource types from schema
	resourceTypes := r.schema.GetResourceTypes()

	for _, resourceType := range resourceTypes {
		if !strings.HasPrefix(resourceType, "azurerm_") {
			continue
		}

		forceNewAttrs := r.schema.GetForceNewAttributes(resourceType)
		if len(forceNewAttrs) == 0 {
			continue
		}

		// Build schema for the ForceNew attributes
		attrSchemas := make([]hclext.AttributeSchema, len(forceNewAttrs))
		for i, attr := range forceNewAttrs {
			attrSchemas[i] = hclext.AttributeSchema{Name: attr}
		}
		bodySchema := &hclext.BodySchema{Attributes: attrSchemas}

		// Get old and new content for this resource type
		oldContent, err := runner.GetOldResourceContent(resourceType, bodySchema, nil)
		if err != nil {
			return fmt.Errorf("get old %s: %w", resourceType, err)
		}
		newContent, err := runner.GetNewResourceContent(resourceType, bodySchema, nil)
		if err != nil {
			return fmt.Errorf("get new %s: %w", resourceType, err)
		}

		// Build map of old resources by name
		oldByName := make(map[string]*hclext.Block)
		for _, block := range oldContent.Blocks {
			if len(block.Labels) >= 2 {
				oldByName[block.Labels[1]] = block
			}
		}

		// Compare each new resource to its old version
		for _, newBlock := range newContent.Blocks {
			if len(newBlock.Labels) < 2 {
				continue
			}
			name := newBlock.Labels[1]
			oldBlock, exists := oldByName[name]
			if !exists {
				continue // New resource, not a ForceNew change
			}

			// Check each ForceNew attribute
			for _, attr := range forceNewAttrs {
				var oldAttr, newAttr *hclext.Attribute
				if oldBlock.Body != nil {
					oldAttr = oldBlock.Body.Attributes[attr]
				}
				if newBlock.Body != nil {
					newAttr = newBlock.Body.Attributes[attr]
				}

				changed, oldVal, newVal := r.attributeChanged(oldAttr, newAttr)
				if changed {
					// Include remediation in message per CR-0002
					message := fmt.Sprintf(
						"Changing %q forces recreation of %s.%s (old: %s, new: %s). "+
							"Consider using a moved block or creating a new resource with a different name.",
						attr, resourceType, name, formatValue(oldVal), formatValue(newVal),
					)
					issueRange := hcl.Range{}
					if newAttr != nil {
						issueRange = newAttr.Range
					} else if newBlock.DefRange != (hcl.Range{}) {
						issueRange = newBlock.DefRange
					}
					if err := runner.EmitIssue(r, message, issueRange); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

// attributeChanged checks if an attribute value has changed.
// Returns whether changed, old value string, new value string.
func (r *AzurermForceNewRule) attributeChanged(oldAttr, newAttr *hclext.Attribute) (bool, string, string) {
	// Both nil = no change
	if oldAttr == nil && newAttr == nil {
		return false, "", ""
	}
	// One nil, one not = change
	if oldAttr == nil {
		newVal := evalAttr(newAttr)
		return true, "<not set>", newVal
	}
	if newAttr == nil {
		oldVal := evalAttr(oldAttr)
		return true, oldVal, "<not set>"
	}

	// Both present - compare evaluated values
	oldVal := evalAttr(oldAttr)
	newVal := evalAttr(newAttr)
	return oldVal != newVal, oldVal, newVal
}

// evalAttr evaluates an HCL attribute to a string representation.
// Supports both direct expression evaluation (local runner) and pre-evaluated Value (gRPC).
func evalAttr(attr *hclext.Attribute) string {
	if attr == nil {
		return "<not set>"
	}

	// First check if we have a pre-evaluated Value (from gRPC serialization)
	if attr.Value != cty.NilVal && !attr.Value.IsNull() {
		return formatCtyValue(attr.Value)
	}

	// Fall back to expression evaluation (direct runner, not gRPC)
	if attr.Expr != nil {
		val, diags := attr.Expr.Value(nil)
		if !diags.HasErrors() {
			return formatCtyValue(val)
		}
	}

	return "<dynamic>"
}

// formatCtyValue formats a cty.Value for display.
func formatCtyValue(val cty.Value) string {
	if val.IsNull() {
		return "<null>"
	}
	if !val.IsKnown() {
		return "<unknown>"
	}

	switch val.Type() {
	case cty.String:
		return val.AsString()
	case cty.Number:
		bf := val.AsBigFloat()
		return bf.Text('f', -1)
	case cty.Bool:
		if val.True() {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%#v", val)
	}
}

// formatValue formats a value for display in messages.
func formatValue(v string) string {
	if v == "" {
		return "<not set>"
	}
	return v
}
