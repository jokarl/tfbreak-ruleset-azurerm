// Package rules provides the tfbreak rules for Azure RM provider.
// This package follows the tflint-ruleset-azurerm pattern of organizing rules.
package rules

import (
	"github.com/jokarl/tfbreak-plugin-sdk/tflint"
)

// Rules is the list of all rules in this ruleset.
// This follows the tflint-ruleset-azurerm pattern.
var Rules = []tflint.Rule{
	NewAzurermForceNewRule(),
}
