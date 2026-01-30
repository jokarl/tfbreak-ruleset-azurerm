# tfbreak-ruleset-azurerm

[![CI](https://github.com/jokarl/tfbreak-ruleset-azurerm/actions/workflows/ci.yml/badge.svg)](https://github.com/jokarl/tfbreak-ruleset-azurerm/actions/workflows/ci.yml)
[![Release](https://github.com/jokarl/tfbreak-ruleset-azurerm/actions/workflows/release.yml/badge.svg)](https://github.com/jokarl/tfbreak-ruleset-azurerm/releases)

tfbreak plugin for Azure RM provider. Detects breaking changes caused by ForceNew attributes in Terraform configurations.

## Requirements

- tfbreak v0.1.0 or later
- Go 1.21+ (for building from source)

## What This Plugin Does

When you modify certain attributes in Azure RM resources, Terraform will destroy and recreate the resource instead of updating it in place. This is known as a "ForceNew" change. This plugin detects these breaking changes by:

1. Comparing your old and new Terraform configurations
2. Identifying changes to attributes marked as ForceNew in the Azure RM provider schema
3. Reporting these as errors so you can review them before applying

For example, changing the `location` of an `azurerm_resource_group` will destroy the entire resource group and all resources within it.

## Installation

### Plugin Discovery (Recommended)

Add the following to your `.tfbreak.hcl`:

```hcl
plugin "azurerm" {
    enabled = true
    version = "0.1.0"
    source  = "github.com/jokarl/tfbreak-ruleset-azurerm"
}
```

### Manual Installation

Download the appropriate binary for your platform from the [releases page](https://github.com/jokarl/tfbreak-ruleset-azurerm/releases) and place it in:

- **Linux/macOS:** `~/.tfbreak.d/plugins/`
- **Windows:** `%USERPROFILE%\.tfbreak.d\plugins\`

The binary should be named `tfbreak-ruleset-azurerm` (or `tfbreak-ruleset-azurerm.exe` on Windows).

## Configuration

### Basic Configuration

```hcl
plugin "azurerm" {
    enabled = true
}
```

### Rule Configuration

Individual rules can be enabled or disabled:

```hcl
plugin "azurerm" {
    enabled = true
}

rule "azurerm_force_new" {
    enabled = true  # This is the default
}
```

## Rules

See the [Rules documentation](docs/README.md) for a complete list of rules and their descriptions.

| Rule | Description | Default |
|------|-------------|---------|
| [azurerm_force_new](docs/rules/azurerm_force_new.md) | Detects ForceNew attribute changes | Enabled |

## Example Output

```
$ tfbreak check --old main-old.tf --new main-new.tf

Error: Changing "location" forces recreation of azurerm_resource_group.example (azurerm_force_new)

  on main-new.tf line 3:
   3:     location = "eastus"

Consider using a moved block or creating a new resource with a different name.

Found 1 breaking change(s).
```

## Building from Source

### Prerequisites

- Go 1.21 or later
- Git

### Build

```bash
# Clone the repository
git clone https://github.com/jokarl/tfbreak-ruleset-azurerm.git
cd tfbreak-ruleset-azurerm

# Build the plugin
go build -o tfbreak-ruleset-azurerm .

# Install to plugin directory
mkdir -p ~/.tfbreak.d/plugins
cp tfbreak-ruleset-azurerm ~/.tfbreak.d/plugins/
```

### Running Tests

```bash
go test -race -v ./...
```

## Schema Updates

This plugin embeds the Azure RM provider schema to detect ForceNew attributes. The schema is updated weekly via GitHub Actions. See [Schema Documentation](docs/schema.md) for details on how this works and its limitations.

## Contributing

Contributions are welcome! Please see:

- [Development Guide](docs/development.md) for building and testing
- [Architecture Decision Records](docs/adr/) for design decisions

## Related Projects

- [tfbreak](https://github.com/jokarl/tfbreak-core) - The main tfbreak CLI
- [tfbreak-plugin-sdk](https://github.com/jokarl/tfbreak-plugin-sdk) - SDK for building tfbreak plugins
- [tflint-ruleset-azurerm](https://github.com/terraform-linters/tflint-ruleset-azurerm) - TFLint plugin for Azure RM (inspiration for project structure)

## License

Mozilla Public License 2.0. See [LICENSE](LICENSE) for details.
