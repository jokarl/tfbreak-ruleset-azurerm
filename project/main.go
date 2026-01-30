// Package project provides version and metadata for tfbreak-ruleset-azurerm.
// This package follows the tflint-ruleset-azurerm pattern.
package project

import "fmt"

// Version is the ruleset version.
const Version string = "0.1.0"

// ReferenceLink returns the documentation link for a rule.
func ReferenceLink(name string) string {
	return fmt.Sprintf("https://github.com/jokarl/tfbreak-ruleset-azurerm/blob/v%s/docs/rules/%s.md", Version, name)
}
