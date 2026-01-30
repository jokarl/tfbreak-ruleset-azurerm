package rules

import (
	"testing"
)

func TestRules_Contains(t *testing.T) {
	// Verify the Rules slice contains the expected rules
	if len(Rules) == 0 {
		t.Fatal("Rules slice is empty")
	}

	// Check that azurerm_force_new rule is present
	found := false
	for _, rule := range Rules {
		if rule.Name() == "azurerm_force_new" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Rules slice does not contain azurerm_force_new rule")
	}
}

func TestRules_UniqueNames(t *testing.T) {
	// Verify all rule names are unique
	names := make(map[string]bool)
	for _, rule := range Rules {
		name := rule.Name()
		if names[name] {
			t.Errorf("Duplicate rule name: %s", name)
		}
		names[name] = true
	}
}
