// Package main provides a tool to extract the Azure RM provider schema.
// This tool extracts the provider schema from a Terraform installation and
// converts it to a format suitable for embedding in the plugin.
//
// Usage:
//
//	go run ./tools/extract-schema -output schema/azurerm.json.gz
//
// The tool requires the azurerm provider to be installed. You can install it by:
//
//  1. Creating a minimal Terraform configuration:
//     terraform { required_providers { azurerm = { source = "hashicorp/azurerm" } } }
//
//  2. Running terraform init
//
//  3. Running this tool from the same directory
package main

import (
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
)

// ProviderSchemaOutput represents the output of `terraform providers schema -json`.
type ProviderSchemaOutput struct {
	ProviderSchemas map[string]*ProviderSchema `json:"provider_schemas"`
}

// ProviderSchema represents a single provider's schema.
type ProviderSchema struct {
	ResourceSchemas   map[string]*ResourceSchema   `json:"resource_schemas"`
	DataSourceSchemas map[string]*ResourceSchema   `json:"data_source_schemas"`
}

// ResourceSchema represents the schema for a resource or data source.
type ResourceSchema struct {
	Block *BlockSchema `json:"block"`
}

// BlockSchema represents a block within a schema.
type BlockSchema struct {
	Attributes map[string]*AttributeSchema   `json:"attributes,omitempty"`
	BlockTypes map[string]*NestedBlockSchema `json:"block_types,omitempty"`
}

// AttributeSchema represents an attribute.
type AttributeSchema struct {
	Type        interface{} `json:"type,omitempty"`
	Description string      `json:"description,omitempty"`
	Required    bool        `json:"required,omitempty"`
	Optional    bool        `json:"optional,omitempty"`
	Computed    bool        `json:"computed,omitempty"`
	ForceNew    bool        `json:"force_new,omitempty"`
	Sensitive   bool        `json:"sensitive,omitempty"`
	Deprecated  string      `json:"deprecated,omitempty"`
}

// NestedBlockSchema represents a nested block.
type NestedBlockSchema struct {
	NestingMode string       `json:"nesting_mode"`
	Block       *BlockSchema `json:"block"`
	MinItems    int          `json:"min_items,omitempty"`
	MaxItems    int          `json:"max_items,omitempty"`
}

// OutputSchema is the simplified schema format we embed in the plugin.
type OutputSchema struct {
	ResourceSchemas map[string]*ResourceSchema `json:"resource_schemas"`
}

func main() {
	output := flag.String("output", "azurerm.json.gz", "Output file path (will be gzip compressed)")
	providerKey := flag.String("provider", "registry.terraform.io/hashicorp/azurerm", "Provider key in the schema output")
	flag.Parse()

	// Run terraform providers schema
	fmt.Println("Running terraform providers schema -json...")
	cmd := exec.Command("terraform", "providers", "schema", "-json")
	cmdOutput, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			fmt.Fprintf(os.Stderr, "terraform providers schema failed:\n%s\n", string(exitErr.Stderr))
		}
		fmt.Fprintf(os.Stderr, "Error running terraform: %v\n", err)
		fmt.Fprintln(os.Stderr, "\nMake sure you have:")
		fmt.Fprintln(os.Stderr, "1. Terraform installed")
		fmt.Fprintln(os.Stderr, "2. A configuration with azurerm provider")
		fmt.Fprintln(os.Stderr, "3. Run 'terraform init' first")
		os.Exit(1)
	}

	// Parse the schema
	var fullSchema ProviderSchemaOutput
	if err := json.Unmarshal(cmdOutput, &fullSchema); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing terraform schema: %v\n", err)
		os.Exit(1)
	}

	// Extract azurerm provider schema
	providerSchema, ok := fullSchema.ProviderSchemas[*providerKey]
	if !ok {
		fmt.Fprintf(os.Stderr, "Provider %s not found in schema\n", *providerKey)
		fmt.Fprintln(os.Stderr, "Available providers:")
		for key := range fullSchema.ProviderSchemas {
			fmt.Fprintf(os.Stderr, "  - %s\n", key)
		}
		os.Exit(1)
	}

	// Create output schema
	outputSchema := OutputSchema{
		ResourceSchemas: providerSchema.ResourceSchemas,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(outputSchema)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling schema: %v\n", err)
		os.Exit(1)
	}

	// Write gzipped output
	f, err := os.Create(*output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	gzw := gzip.NewWriter(f)
	defer gzw.Close()

	if _, err := gzw.Write(jsonData); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing compressed data: %v\n", err)
		os.Exit(1)
	}

	// Print stats
	resourceCount := len(outputSchema.ResourceSchemas)
	forceNewCount := 0
	for _, rs := range outputSchema.ResourceSchemas {
		forceNewCount += countForceNew(rs.Block)
	}

	fmt.Printf("Extracted schema for %d resources with %d ForceNew attributes\n", resourceCount, forceNewCount)
	fmt.Printf("Written to %s\n", *output)
}

func countForceNew(block *BlockSchema) int {
	if block == nil {
		return 0
	}
	count := 0
	for _, attr := range block.Attributes {
		if attr.ForceNew {
			count++
		}
	}
	for _, nested := range block.BlockTypes {
		count += countForceNew(nested.Block)
	}
	return count
}
