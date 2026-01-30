// Package main is the entry point for tfbreak-ruleset-azurerm.
// This follows the tflint-ruleset-azurerm pattern of using plugin.Serve().
package main

import (
	"github.com/jokarl/tfbreak-plugin-sdk/plugin"
	"github.com/jokarl/tfbreak-plugin-sdk/tflint"
	"github.com/jokarl/tfbreak-ruleset-azurerm/project"
	"github.com/jokarl/tfbreak-ruleset-azurerm/rules"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		RuleSet: &tflint.BuiltinRuleSet{
			Name:    "azurerm",
			Version: project.Version,
			Rules:   rules.Rules,
		},
	})
}
