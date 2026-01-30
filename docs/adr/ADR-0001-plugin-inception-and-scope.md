---
status: accepted
date: 2026-01-29
decision-makers: [project maintainers]
consulted: []
informed: []
---

# tfbreak-ruleset-azurerm: Plugin Inception and Scope

## Context and Problem Statement

tfbreak-core provides provider-agnostic breaking change detection. Azure Resource Manager (azurerm) resources have provider-specific behaviors that cannot be detected by generic rules:

1. Many resource attributes are marked with "Changing this forces a new resource to be created"
2. Azure-specific naming constraints
3. Provider-specific lifecycle behaviors

How should tfbreak-ruleset-azurerm implement Azure-specific detection while:
- Adhering to tfbreak's tflint-aligned minimal plugin interface
- Aligning with tflint-ruleset-azurerm patterns where appropriate

## Decision Drivers

* Must implement tfbreak's RuleSet interface (tflint-aligned)
* Must use Runner interface to access old/new configs
* Must report findings via EmitIssue
* Should align with tflint-ruleset-azurerm project structure
* Plugin decides detection approach - not forced by interface
* Must scale to 900+ azurerm resource types
* Must minimize maintenance burden

## Considered Options

* **Option 1: Schema-driven detection with tflint-aligned project structure**
* **Option 2: Per-resource rules (like tflint-ruleset-azurerm)**
* **Option 3: Hybrid approach**

## Decision Outcome

Chosen option: "Option 1: Schema-driven detection with tflint-aligned project structure", because it automatically covers all 900+ resources while adhering to the minimal plugin interface and following tflint-ruleset-azurerm's project organization patterns.

### What We Align With (from tflint-ruleset-azurerm)

| Pattern | tflint-ruleset-azurerm | Our Approach |
|---------|------------------------|--------------|
| Entry point | `main.go` with `plugin.Serve()` | Same |
| Version management | `project/main.go` with `Version` const | Same |
| Rule organization | `rules/` directory | Same |
| Rule listing | `rules/provider.go` with `Rules` var | Same |
| Severity levels | ERROR, WARNING, NOTICE | Same |

### What We Intentionally Differ On

| Pattern | tflint-ruleset-azurerm | Our Approach | Reason |
|---------|------------------------|--------------|--------|
| Rule structure | 150+ individual rules | Single generic rule | Schema-driven detection |
| Code generation | Auto-generate from API specs | Embed provider schema | ForceNew is already in schema |
| Detection scope | Static config validation | Old vs new comparison | Different problem domain |

### Schema Version Strategy

The plugin embeds a single provider schema version (typically the latest). This creates a potential mismatch when users run the plugin against repositories using older provider versions, since ForceNew behavior can change between versions.

**Considered alternatives:**
1. **Runtime schema extraction** - Extract schema from the repo's actual provider at runtime. Most accurate but requires terraform binary, provider download, and adds latency.
2. **Bundle multiple schemas** - Ship schemas for major versions and select based on the repo's version constraint. More accurate but increases binary size and maintenance.
3. **External schema service** - Fetch schemas from a CDN/API. Accurate but requires network access and external infrastructure.
4. **Single embedded schema** - Ship one schema version, document the limitation.

**Decision:** Single embedded schema (option 4), because:
- Simplicity aligns with MVP goals
- Matches tflint-ruleset-azurerm's approach
- Avoids runtime complexity and external dependencies
- Users typically keep plugins updated alongside providers

**Limitation:** ForceNew detection may be inaccurate for repositories using provider versions that differ significantly from the embedded schema version. Users should keep their plugin version aligned with their provider version for best accuracy.

### Consequences

* Good, because project structure is familiar to tflint plugin developers
* Good, because one rule covers all resources automatically
* Good, because schema updates = rule coverage updates
* Good, because adheres to tflint-aligned interface
* Bad, because plugin releases lag behind provider releases
* Neutral, because requires schema extraction tooling (instead of code generation)

### Confirmation

1. Successfully implement RuleSet interface with tflint-aligned structure
2. Demonstrate ForceNew detection across multiple resource types
3. Validate that schema-driven approach works with minimal interface

## Pros and Cons of the Options

### Option 1: Schema-driven detection with tflint-aligned structure

Adopt tflint-ruleset-azurerm's project structure (main.go, project/, rules/) while using schema-driven detection internally.

```go
// main.go (tflint-aligned)
func main() {
    plugin.Serve(&plugin.ServeOpts{
        RuleSet: &plugin.BuiltinRuleSet{
            Name:    "azurerm",
            Version: project.Version,
            Rules:   rules.Rules,
        },
    })
}

// rules/provider.go (tflint-aligned)
var Rules = []plugin.Rule{
    NewAzurermForceNewRule(),
}

// rules/azurerm_force_new.go - schema-driven detection
func (r *AzurermForceNewRule) Check(runner plugin.Runner) error {
    oldResources, _ := runner.GetOldResourceContent(...)
    newResources, _ := runner.GetNewResourceContent(...)

    for resourceType, resources := range newResources {
        forceNewAttrs := r.schema.GetForceNewAttributes(resourceType)
        // ... compare and detect ...
        runner.EmitIssue(r, "Changing 'location' forces recreation", range)
    }
    return nil
}
```

* Good, because familiar project structure for tflint developers
* Good, because one rule handles all 900+ resources
* Good, because no per-resource code generation needed
* Good, because fits within minimal interface

### Option 2: Per-resource rules (like tflint-ruleset-azurerm)

Create individual rules for each resource type, potentially auto-generated.

```go
// rules/azurerm_resource_group_force_new_location.go
type AzurermResourceGroupForceNewLocationRule struct {
    plugin.DefaultRule
}

func (r *AzurermResourceGroupForceNewLocationRule) Check(runner plugin.Runner) error {
    // Check specific attribute for specific resource
}
```

* Bad, because 900+ rule implementations needed (or 900+ generated files)
* Bad, because requires code generation infrastructure
* Bad, because maintenance burden for each provider release

### Option 3: Hybrid approach

* Good, because flexibility
* Neutral, because complexity of two approaches

## More Information

### Project Structure (tflint-ruleset-azurerm aligned)

```
tfbreak-ruleset-azurerm/
├── main.go                           # Entry point (tflint-aligned)
├── project/
│   └── main.go                       # Version constant (tflint-aligned)
├── rules/
│   ├── provider.go                   # Rules slice (tflint-aligned)
│   └── azurerm_force_new.go          # Single schema-driven rule
├── schema/
│   ├── schema.go                     # Schema loader
│   └── azurerm.json.gz               # Embedded provider schema
├── tools/
│   └── extract-schema/               # Schema extraction tool
└── docs/
    ├── adr/
    └── cr/
```

### Entry Point (tflint-aligned)

```go
// main.go
package main

import (
    "github.com/jokarl/tfbreak-core/plugin"
    "github.com/jokarl/tfbreak-ruleset-azurerm/project"
    "github.com/jokarl/tfbreak-ruleset-azurerm/rules"
)

func main() {
    plugin.Serve(&plugin.ServeOpts{
        RuleSet: &plugin.BuiltinRuleSet{
            Name:    "azurerm",
            Version: project.Version,
            Rules:   rules.Rules,
        },
    })
}
```

### Version Management (tflint-aligned)

```go
// project/main.go
package project

import "fmt"

// Version is ruleset version
const Version string = "0.1.0"

// ReferenceLink returns the rule reference link
func ReferenceLink(name string) string {
    return fmt.Sprintf("https://github.com/jokarl/tfbreak-ruleset-azurerm/blob/v%s/docs/rules/%s.md", Version, name)
}
```

### Schema Extraction

Provider schema extracted via `terraform providers schema -json`:

```yaml
# .github/workflows/update-schema.yml
name: Update Provider Schema
on:
  schedule:
    - cron: '0 0 * * 0'  # Weekly
jobs:
  update-schema:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: hashicorp/setup-terraform@v3
      - run: |
          cd tools/extract-schema
          go run . -provider hashicorp/azurerm -version latest \
            -output ../../schema/azurerm.json.gz
      - uses: peter-evans/create-pull-request@v5
```

### Severity Levels (tflint-aligned)

| Condition | Severity | tflint Equivalent | Meaning |
|-----------|----------|-------------------|---------|
| ForceNew attribute changed | ERROR | `tflint.ERROR` | Resource will be destroyed/recreated |
| Deprecated attribute | WARNING | `tflint.WARNING` | Potential issue, review recommended |
| Informational | NOTICE | `tflint.NOTICE` | Informational |

### Phased Implementation

**Phase 1: Foundation (tflint-aligned structure)**
- Implement entry point in `main.go`
- Create `project/main.go` with Version
- Create `rules/provider.go` with Rules slice
- Implement basic schema embedding

**Phase 2: Enhancement**
- Single `azurerm_force_new` rule
- Deprecated attribute detection

**Phase 3: Advanced**
- Cross-resource analysis
- Custom configurations

### References

- [tfbreak-core ADR-0002: Plugin Architecture](https://github.com/jokarl/tfbreak-core/docs/adr/ADR-0002-plugin-architecture.md)
- [tflint-ruleset-azurerm](https://github.com/terraform-linters/tflint-ruleset-azurerm) - Reference for project structure
- [Azure RM Provider](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs)
