# Schema Documentation

This document explains how tfbreak-ruleset-azurerm uses the Azure RM provider schema for ForceNew detection.

## Overview

The plugin embeds a compressed JSON schema extracted from the Azure RM Terraform provider. This schema contains metadata about every resource type, including which attributes are marked as ForceNew.

## How It Works

### Schema Extraction

The schema is extracted using Terraform's built-in schema introspection:

```bash
terraform providers schema -json
```

This outputs the complete provider schema, including:
- All resource types
- All data sources
- Attribute metadata (type, required, optional, computed, ForceNew, deprecated)
- Nested block structures

### Schema Processing

The extraction tool (`tools/extract-schema/main.go`):
1. Runs `terraform providers schema -json`
2. Parses the JSON output
3. Extracts only the Azure RM provider's resource schemas
4. Compresses the result with gzip
5. Saves to `schema/azurerm.json.gz`

### Schema Loading

At runtime:
1. The schema is embedded in the binary using Go's `embed` directive
2. On first use, it's decompressed and parsed into memory
3. The parsed schema is cached for subsequent queries

### ForceNew Detection

When checking a resource:
1. Look up the resource type in the schema (e.g., `azurerm_resource_group`)
2. Get all attributes marked with `force_new: true`
3. Compare old and new values for these attributes
4. Report changes as errors

## Schema Version Strategy

### Current Approach: Single Embedded Schema

The plugin embeds a single schema version, typically from the latest Azure RM provider release.

**Advantages:**
- Simple implementation
- No runtime dependencies
- Fast startup (schema already in binary)
- Works offline

**Limitations:**
- Schema may not match the provider version in use
- ForceNew behavior can change between provider versions

### Version Mismatch Scenarios

| Your Provider | Embedded Schema | Result |
|---------------|-----------------|--------|
| Same version | Same version | Accurate detection |
| Older version | Newer version | May flag attributes not ForceNew in your version |
| Newer version | Older version | May miss new ForceNew attributes |

### Recommendations

1. **Keep plugin updated** - Update the plugin when you update your Azure RM provider
2. **Review warnings** - If a flagged attribute seems incorrect, check the provider documentation
3. **Pin versions** - Consider pinning both provider and plugin versions together

## Schema Updates

### Automated Updates

A GitHub Action runs weekly to update the schema:

1. Creates a temporary Terraform configuration with the latest Azure RM provider
2. Runs `terraform init` and `terraform providers schema -json`
3. Extracts and compresses the schema
4. Creates a pull request if the schema changed

See `.github/workflows/update-schema.yml` for the workflow definition.

### Manual Updates

To update the schema manually:

```bash
# Create a temporary directory
mkdir /tmp/azurerm-schema && cd /tmp/azurerm-schema

# Create a Terraform configuration
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

# Initialize Terraform
terraform init

# Extract the schema
cd /path/to/tfbreak-ruleset-azurerm
go run ./tools/extract-schema -output schema/azurerm.json.gz
```

### Specific Provider Versions

To extract a schema from a specific provider version:

```bash
# In your temporary Terraform configuration
cat > main.tf << 'EOF'
terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "= 3.100.0"  # Specific version
    }
  }
}
EOF
```

## Schema Structure

### Overview

```json
{
  "resource_schemas": {
    "azurerm_resource_group": {
      "block": {
        "attributes": {
          "name": {
            "type": "string",
            "required": true,
            "force_new": true
          },
          "location": {
            "type": "string",
            "required": true,
            "force_new": true
          },
          "tags": {
            "type": ["map", "string"],
            "optional": true
          }
        }
      }
    }
  }
}
```

### Attribute Properties

| Property | Type | Description |
|----------|------|-------------|
| `type` | string/array | HCL type (string, number, bool, list, map, etc.) |
| `required` | bool | Must be specified |
| `optional` | bool | May be specified |
| `computed` | bool | Computed by the provider |
| `force_new` | bool | Changes force resource recreation |
| `sensitive` | bool | Value is sensitive |
| `deprecated` | string | Deprecation message |
| `description` | string | Attribute description |

### Nested Blocks

Resources can have nested blocks (e.g., `identity`, `network_rules`):

```json
{
  "block": {
    "attributes": { ... },
    "block_types": {
      "identity": {
        "nesting_mode": "list",
        "block": {
          "attributes": {
            "type": { "type": "string", "required": true, "force_new": true }
          }
        },
        "max_items": 1
      }
    }
  }
}
```

The schema loader recursively searches nested blocks for ForceNew attributes.

## Alternatives Considered

### Runtime Schema Extraction

**Description:** Extract schema from the user's actual Terraform environment at runtime.

**Pros:**
- Always matches user's provider version
- Most accurate detection

**Cons:**
- Requires `terraform` binary
- Slow (provider download, schema extraction)
- May fail in CI environments

**Decision:** Not implemented for v0.1.0 due to complexity.

### Multiple Embedded Schemas

**Description:** Bundle schemas for multiple provider versions, select based on version constraint.

**Pros:**
- Better version matching
- Still works offline

**Cons:**
- Larger binary size
- Maintenance burden
- Limited version coverage

**Decision:** Not implemented for v0.1.0 due to complexity.

### External Schema Service

**Description:** Fetch schemas from a CDN or API.

**Pros:**
- Always up-to-date
- Small binary size

**Cons:**
- Requires network access
- External dependency
- Latency

**Decision:** Not implemented for v0.1.0 due to infrastructure requirements.

## Future Improvements

Potential future enhancements:

1. **Version-aware schema selection** - Detect provider version and select appropriate schema
2. **Schema caching** - Cache downloaded schemas locally
3. **Hybrid approach** - Embedded schema with optional runtime refresh
4. **Schema diff reports** - Show what changed between schema versions

## Related

- [ADR-0001: Plugin Inception and Scope](adr/ADR-0001-plugin-inception-and-scope.md) - Design decision
- [tools/extract-schema](../tools/extract-schema/) - Schema extraction tool
- [Azure RM Provider](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs)
