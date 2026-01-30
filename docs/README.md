# tfbreak-ruleset-azurerm Rules

This document lists all rules available in the tfbreak-ruleset-azurerm plugin.

## Overview

Unlike tflint-ruleset-azurerm which has individual rules for each validation type, tfbreak-ruleset-azurerm uses a **schema-driven approach** with a single rule that automatically detects ForceNew changes across all 900+ Azure RM resource types.

This approach:
- Automatically covers all resources without per-resource rule maintenance
- Stays up-to-date with provider schema changes via automated updates
- Reduces binary size and complexity

## Rules

| Rule | Description | Severity | Default |
|------|-------------|----------|---------|
| [azurerm_force_new](rules/azurerm_force_new.md) | Detects changes to ForceNew attributes | ERROR | Enabled |

## Severity Levels

| Severity | Meaning |
|----------|---------|
| ERROR | Breaking change that will cause resource destruction/recreation |
| WARNING | Potential issue that should be reviewed |
| NOTICE | Informational finding |

The `azurerm_force_new` rule uses ERROR severity because ForceNew changes always result in resource destruction, which is considered a breaking change.

## Planned Rules

Future versions may include:

| Rule | Description | Status |
|------|-------------|--------|
| azurerm_deprecated | Detects deprecated attributes | Planned |
| azurerm_lifecycle_ignore | Detects lifecycle ignore_changes gaps | Planned |

## Configuration

### Disabling Rules

Rules can be disabled in your `.tfbreak.hcl`:

```hcl
rule "azurerm_force_new" {
    enabled = false
}
```

### Inline Suppression

You can suppress specific findings using tfbreak annotations:

```hcl
resource "azurerm_resource_group" "example" {
    name     = "my-rg"
    # tfbreak-ignore: azurerm_force_new
    location = "eastus"  # Intentional migration
}
```

## See Also

- [Schema Documentation](schema.md) - How ForceNew detection works
- [Development Guide](development.md) - Adding new rules
