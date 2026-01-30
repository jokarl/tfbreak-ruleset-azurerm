package project

import (
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}

	// Version should be semver-like
	parts := strings.Split(Version, ".")
	if len(parts) < 2 {
		t.Errorf("Version %s does not look like semver", Version)
	}
}

func TestReferenceLink(t *testing.T) {
	link := ReferenceLink("azurerm_force_new")

	if link == "" {
		t.Error("ReferenceLink should not return empty string")
	}

	// Should contain the version
	if !strings.Contains(link, Version) {
		t.Errorf("ReferenceLink should contain version %s, got %s", Version, link)
	}

	// Should contain the rule name
	if !strings.Contains(link, "azurerm_force_new") {
		t.Errorf("ReferenceLink should contain rule name, got %s", link)
	}

	// Should be a GitHub URL
	if !strings.HasPrefix(link, "https://github.com/") {
		t.Errorf("ReferenceLink should be a GitHub URL, got %s", link)
	}
}
