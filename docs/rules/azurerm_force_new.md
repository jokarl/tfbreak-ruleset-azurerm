# azurerm_force_new

Detects changes to ForceNew attributes in Azure RM resources.

## Rule Details

| Property | Value |
|----------|-------|
| Rule ID | `azurerm_force_new` |
| Severity | ERROR |
| Enabled by default | Yes |
| Since | v0.1.0 |

## Description

In Terraform, some resource attributes are marked with "Changing this forces a new resource to be created" (ForceNew). When you modify these attributes, Terraform cannot update the existing resource in place. Instead, it must:

1. Destroy the existing resource
2. Create a new resource with the new configuration

This can be a breaking change because:
- Data loss may occur if the resource contains state
- Dependent resources may also be affected
- Service interruption during recreation
- Resource IDs change, breaking external references

## How It Works

This rule uses a **schema-driven detection** approach:

1. Loads the embedded Azure RM provider schema (extracted from `terraform providers schema -json`)
2. For each `azurerm_*` resource in your configuration, retrieves the list of ForceNew attributes from the schema
3. Compares old and new configurations to detect changes to these attributes
4. Reports any ForceNew attribute changes as errors

### Coverage

This single rule automatically covers all 900+ Azure RM resource types. Coverage updates automatically when the embedded schema is updated.

## Examples

### What Gets Flagged

**Changing resource group location:**

```hcl
# Old configuration
resource "azurerm_resource_group" "example" {
    name     = "my-rg"
    location = "westeurope"
}

# New configuration
resource "azurerm_resource_group" "example" {
    name     = "my-rg"
    location = "eastus"  # <- ERROR: ForceNew change detected
}
```

**Changing storage account name:**

```hcl
# Old configuration
resource "azurerm_storage_account" "example" {
    name                     = "mystorageaccount"
    resource_group_name      = azurerm_resource_group.example.name
    location                 = azurerm_resource_group.example.location
    account_tier             = "Standard"
    account_replication_type = "LRS"
}

# New configuration
resource "azurerm_storage_account" "example" {
    name                     = "newstorageaccount"  # <- ERROR: ForceNew change
    resource_group_name      = azurerm_resource_group.example.name
    location                 = azurerm_resource_group.example.location
    account_tier             = "Standard"
    account_replication_type = "LRS"
}
```

### What Does NOT Get Flagged

**Changing non-ForceNew attributes:**

```hcl
# Old configuration
resource "azurerm_resource_group" "example" {
    name     = "my-rg"
    location = "westeurope"
    tags = {
        environment = "dev"
    }
}

# New configuration
resource "azurerm_resource_group" "example" {
    name     = "my-rg"
    location = "westeurope"
    tags = {
        environment = "prod"  # OK: tags is not ForceNew
    }
}
```

**New resources (no old configuration):**

```hcl
# Old configuration: (empty or resource doesn't exist)

# New configuration
resource "azurerm_resource_group" "new_example" {
    name     = "new-rg"
    location = "eastus"  # OK: This is a new resource, not a change
}
```

## How to Suppress

### Using Annotations

Add a `tfbreak-ignore` comment above or on the same line as the attribute:

```hcl
resource "azurerm_resource_group" "example" {
    name     = "my-rg"
    # tfbreak-ignore: azurerm_force_new
    location = "eastus"  # Intentional migration to new region
}
```

Or on the same line:

```hcl
resource "azurerm_resource_group" "example" {
    name     = "my-rg"
    location = "eastus"  # tfbreak-ignore: azurerm_force_new - Migration planned
}
```

### Disabling the Rule

In `.tfbreak.hcl`:

```hcl
rule "azurerm_force_new" {
    enabled = false
}
```

## Remediation Guidance

When you encounter a ForceNew change, consider these options:

### Option 1: Accept the Recreation

If the resource recreation is acceptable (e.g., in a development environment):

```bash
# Review the plan carefully
terraform plan

# Apply if the destruction is acceptable
terraform apply
```

### Option 2: Use a Moved Block

If you're renaming or restructuring but keeping the same physical resource:

```hcl
moved {
    from = azurerm_resource_group.old_name
    to   = azurerm_resource_group.new_name
}
```

### Option 3: Create a New Resource

Create a new resource with a different name, migrate workloads, then remove the old one:

```hcl
# Keep the old resource
resource "azurerm_resource_group" "example" {
    name     = "my-rg"
    location = "westeurope"
}

# Add a new resource in the new location
resource "azurerm_resource_group" "example_v2" {
    name     = "my-rg-v2"
    location = "eastus"
}
```

### Option 4: Import Existing Resource

If the resource already exists with the new configuration:

```bash
terraform import azurerm_resource_group.example /subscriptions/.../resourceGroups/my-rg
```

## Common ForceNew Attributes

Here are some commonly encountered ForceNew attributes by resource type:

| Resource | ForceNew Attributes |
|----------|---------------------|
| `azurerm_resource_group` | `name`, `location` |
| `azurerm_storage_account` | `name`, `resource_group_name`, `location`, `account_kind` |
| `azurerm_virtual_network` | `name`, `resource_group_name`, `location` |
| `azurerm_subnet` | `name`, `resource_group_name`, `virtual_network_name` |
| `azurerm_virtual_machine` | `name`, `resource_group_name`, `location`, `availability_set_id` |

For a complete list, refer to the [Azure RM Provider documentation](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs) - attributes marked with "Changing this forces a new resource to be created" are ForceNew.

## Related

- [Schema Documentation](../schema.md) - How schema-driven detection works
- [Azure RM Provider Docs](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs)
