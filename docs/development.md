# Development Guide

This guide covers how to build, test, and contribute to tfbreak-ruleset-azurerm.

## Prerequisites

- Go 1.21 or later
- Git
- Terraform (for schema extraction)

## Project Structure

```
tfbreak-ruleset-azurerm/
├── main.go                    # Plugin entry point
├── go.mod                     # Go module definition
├── project/
│   └── main.go                # Version and metadata
├── rules/
│   ├── provider.go            # Rules registry
│   ├── azurerm_force_new.go   # ForceNew detection rule
│   └── *_test.go              # Rule tests
├── schema/
│   ├── schema.go              # Schema loader
│   ├── schema_test.go         # Schema tests
│   └── azurerm.json.gz        # Embedded provider schema
├── tools/
│   └── extract-schema/
│       └── main.go            # Schema extraction tool
├── docs/
│   ├── README.md              # Rules documentation index
│   ├── schema.md              # Schema documentation
│   ├── development.md         # This file
│   ├── rules/                 # Individual rule documentation
│   ├── adr/                   # Architecture Decision Records
│   └── cr/                    # Change Requests
└── .github/
    └── workflows/
        ├── ci.yml             # CI workflow
        ├── release.yml        # Release workflow
        └── update-schema.yml  # Schema update workflow
```

## Building

### Standard Build

```bash
go build -o tfbreak-ruleset-azurerm .
```

### With Version Information

```bash
VERSION=$(git describe --tags --always --dirty)
COMMIT=$(git rev-parse HEAD)
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

go build -trimpath \
    -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
    -o tfbreak-ruleset-azurerm .
```

### Cross-Compilation

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o tfbreak-ruleset-azurerm-linux-amd64 .

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o tfbreak-ruleset-azurerm-linux-arm64 .

# macOS AMD64
GOOS=darwin GOARCH=amd64 go build -o tfbreak-ruleset-azurerm-darwin-amd64 .

# macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o tfbreak-ruleset-azurerm-darwin-arm64 .

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o tfbreak-ruleset-azurerm-windows-amd64.exe .
```

## Testing

### Run All Tests

```bash
go test -v ./...
```

### With Race Detection

```bash
go test -race -v ./...
```

### With Coverage

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Run Specific Tests

```bash
# Run tests in a specific package
go test -v ./rules/...

# Run a specific test
go test -v -run TestForceNew_DetectsChange ./rules/...
```

## Code Quality

### Linting

```bash
go vet ./...
```

### Formatting

```bash
go fmt ./...
```

### Vulnerability Check

```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

## Schema Extraction

### Update to Latest Provider

```bash
# Create temporary directory
mkdir -p /tmp/azurerm-schema
cd /tmp/azurerm-schema

# Create Terraform config
cat > main.tf << 'EOF'
terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = ">= 4.0"
    }
  }
}
EOF

# Initialize and extract
terraform init
go run /path/to/tfbreak-ruleset-azurerm/tools/extract-schema -output /path/to/tfbreak-ruleset-azurerm/schema/azurerm.json.gz
```

### Extract Specific Version

```bash
# In main.tf, use exact version:
version = "= 4.5.0"
```

## Adding a New Rule

### 1. Create the Rule File

Create `rules/azurerm_<rule_name>.go`:

```go
package rules

import (
    "github.com/hashicorp/hcl/v2"
    "github.com/jokarl/tfbreak-plugin-sdk/tflint"
    "github.com/jokarl/tfbreak-ruleset-azurerm/project"
)

// AzurermRuleNameRule detects <description>.
type AzurermRuleNameRule struct {
    tflint.DefaultRule
}

// NewAzurermRuleNameRule creates a new rule.
func NewAzurermRuleNameRule() *AzurermRuleNameRule {
    return &AzurermRuleNameRule{}
}

// Name returns the rule name.
func (r *AzurermRuleNameRule) Name() string {
    return "azurerm_rule_name"
}

// Enabled returns whether the rule is enabled by default.
func (r *AzurermRuleNameRule) Enabled() bool {
    return true
}

// Severity returns the rule severity.
func (r *AzurermRuleNameRule) Severity() tflint.Severity {
    return tflint.WARNING  // or tflint.ERROR, tflint.NOTICE
}

// Link returns the documentation link.
func (r *AzurermRuleNameRule) Link() string {
    return project.ReferenceLink(r.Name())
}

// Check implements the rule logic.
func (r *AzurermRuleNameRule) Check(runner tflint.Runner) error {
    // Implementation here
    return nil
}
```

### 2. Register the Rule

Add to `rules/provider.go`:

```go
var Rules = []tflint.Rule{
    NewAzurermForceNewRule(),
    NewAzurermRuleNameRule(),  // Add new rule
}
```

### 3. Write Tests

Create `rules/azurerm_<rule_name>_test.go`:

```go
package rules

import (
    "testing"

    "github.com/jokarl/tfbreak-plugin-sdk/helper"
    "github.com/jokarl/tfbreak-plugin-sdk/tflint"
)

func TestRuleName_DetectsIssue(t *testing.T) {
    rule := NewAzurermRuleNameRule()

    runner := helper.TestRunner(t,
        map[string]string{
            "main.tf": `
resource "azurerm_resource_group" "old" {
    name     = "my-rg"
    location = "westeurope"
}`,
        },
        map[string]string{
            "main.tf": `
resource "azurerm_resource_group" "new" {
    name     = "my-rg"
    location = "eastus"
}`,
        },
    )

    err := rule.Check(runner)
    if err != nil {
        t.Fatalf("Check returned error: %v", err)
    }

    // Assert expected issues
    if len(runner.Issues) == 0 {
        t.Error("Expected issue to be emitted")
    }
}
```

### 4. Document the Rule

Create `docs/rules/azurerm_<rule_name>.md`:

```markdown
# azurerm_rule_name

<Description>

## Rule Details

| Property | Value |
|----------|-------|
| Rule ID | `azurerm_rule_name` |
| Severity | WARNING |
| Enabled by default | Yes |
| Since | v0.2.0 |

## Description

<Detailed description>

## Examples

### What Gets Flagged

<Examples>

### What Does NOT Get Flagged

<Examples>

## How to Suppress

<Suppression instructions>

## Remediation Guidance

<How to fix>
```

### 5. Update Documentation Index

Add to `docs/README.md`:

```markdown
| [azurerm_rule_name](rules/azurerm_rule_name.md) | Description | Enabled |
```

## Local Development with tfbreak

### Using Replace Directive

In your tfbreak-core project, add a replace directive to use your local plugin:

```go
// go.mod
replace github.com/jokarl/tfbreak-ruleset-azurerm => ../tfbreak-ruleset-azurerm
```

### Manual Installation

```bash
# Build the plugin
cd /path/to/tfbreak-ruleset-azurerm
go build -o tfbreak-ruleset-azurerm .

# Install to plugin directory
mkdir -p ~/.tfbreak.d/plugins
cp tfbreak-ruleset-azurerm ~/.tfbreak.d/plugins/
```

## Debugging

### Enable Verbose Output

```bash
# When running tfbreak with the plugin
TFBREAK_LOG=DEBUG tfbreak check ...
```

### Debug Tests

```bash
# Run with verbose output
go test -v -run TestForceNew_DetectsChange ./rules/...

# Debug with delve
dlv test ./rules/ -- -test.run TestForceNew_DetectsChange
```

## Release Process

Releases are automated via GitHub Actions:

1. Merge changes to `main`
2. Release-please creates a release PR
3. Merge the release PR
4. Binaries are automatically built and attached to the release

### Version Bumping

Update `project/main.go`:

```go
const Version string = "0.2.0"
```

## Related Documentation

- [Schema Documentation](schema.md) - How the embedded schema works
- [Rules Documentation](README.md) - Complete rule list
- [ADR-0001](adr/ADR-0001-plugin-inception-and-scope.md) - Architecture decisions
