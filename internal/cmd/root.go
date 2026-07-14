// Package cmd provides the root command and CLI structure.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/EdgarOrtegaRamirez/jsonlforge/internal/cmd/aggregate"
	"github.com/EdgarOrtegaRamirez/jsonlforge/internal/cmd/convert"
	"github.com/EdgarOrtegaRamirez/jsonlforge/internal/cmd/dedup"
	"github.com/EdgarOrtegaRamirez/jsonlforge/internal/cmd/filter"
	"github.com/EdgarOrtegaRamirez/jsonlforge/internal/cmd/flatten"
	"github.com/EdgarOrtegaRamirez/jsonlforge/internal/cmd/merge"
	"github.com/EdgarOrtegaRamirez/jsonlforge/internal/cmd/query"
	"github.com/EdgarOrtegaRamirez/jsonlforge/internal/cmd/schema"
	"github.com/EdgarOrtegaRamirez/jsonlforge/internal/cmd/sort"
	"github.com/EdgarOrtegaRamirez/jsonlforge/internal/cmd/stats"
	"github.com/EdgarOrtegaRamirez/jsonlforge/internal/cmd/transform"
	"github.com/EdgarOrtegaRamirez/jsonlforge/internal/cmd/validate"
)

// RootCmd is the root command for jsonlforge.
var RootCmd = &cobra.Command{
	Use:   "jsonlforge",
	Short: "Comprehensive JSONL processing toolkit",
	Long: `JsonlForge is a fast, powerful CLI toolkit for processing JSONL (JSON Lines) files.

It provides a set of composable commands for filtering, transforming, querying,
validating, and analyzing JSONL data — all designed for speed, streaming efficiency,
and shell pipeline integration.

Commands:
  validate    Validate JSONL file structure and optional JSON Schema
  query       Query JSONL files with jq-like expression syntax
  filter      Filter records by field conditions
  transform   Transform records with field mapping, renaming, and deletion
  flatten     Flatten nested JSON objects into flat key-value pairs
  aggregate   Aggregate records by key with count, sum, avg, min, max, etc.
  schema      Auto-detect or validate against JSON Schema
  sort        Sort records by field value
  dedup       Remove duplicate records
  merge       Merge multiple JSONL files by key
  convert     Convert JSONL to CSV, TSV, or other formats
  stats       Compute summary statistics for fields

Examples:
  jsonlforge validate data.jsonl
  jsonlforge query --fields 'name,age' data.jsonl
  jsonlforge filter --field 'age > 18' data.jsonl
  jsonlforge aggregate --key 'city' --agg 'count:*,sum:salary' data.jsonl
  jsonlforge stats data.jsonl
  jsonlforge convert --to csv data.jsonl > data.csv
  jsonlforge flatten data.jsonl
  jsonlforge dedup data.jsonl
  jsonlforge merge --key 'id' file1.jsonl file2.jsonl

Use "jsonlforge <command> --help" for more information about a command.`,
}

func init() {
	RootCmd.AddCommand(validate.NewCommand())
	RootCmd.AddCommand(query.NewCommand())
	RootCmd.AddCommand(filter.NewCommand())
	RootCmd.AddCommand(transform.NewCommand())
	RootCmd.AddCommand(flatten.NewCommand())
	RootCmd.AddCommand(aggregate.NewCommand())
	RootCmd.AddCommand(schema.NewCommand())
	RootCmd.AddCommand(sort.NewCommand())
	RootCmd.AddCommand(dedup.NewCommand())
	RootCmd.AddCommand(merge.NewCommand())
	RootCmd.AddCommand(convert.NewCommand())
	RootCmd.AddCommand(stats.NewCommand())
}

// Execute runs the root command.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
