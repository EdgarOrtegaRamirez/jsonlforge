// Package dedup provides the "dedup" command for removing duplicate JSONL records.
package dedup

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/EdgarOrtegaRamirez/jsonlforge/internal/pkg/jsonl"
)

// NewCommand creates the dedup command.
func NewCommand() *cobra.Command {
	var (
		key     string
		allKeys bool
		strict  bool
	)

	cmd := &cobra.Command{
		Use:   "dedup [file]",
		Short: "Remove duplicate records from JSONL file",
		Long: `Remove duplicate records from a JSONL file.

By default, dedup removes exact duplicate lines. Use --key to deduplicate
based on a specific field value. Use --all-keys to consider all fields.
Use --strict for exact byte-level comparison.

Examples:
  jsonlforge dedup data.jsonl
  jsonlforge dedup --key 'id' data.jsonl
  jsonlforge dedup --all-keys data.jsonl
  jsonlforge dedup --strict data.jsonl`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := "-"
			if len(args) > 0 && args[0] != "" {
				input = args[0]
			}
			return runDedup(input, key, allKeys, strict)
		},
	}

	cmd.Flags().StringVar(&key, "key", "", "Field to deduplicate by")
	cmd.Flags().BoolVar(&allKeys, "all-keys", false, "Deduplicate by all fields")
	cmd.Flags().BoolVar(&strict, "strict", false, "Exact byte-level deduplication")

	return cmd
}

func runDedup(input, key string, allKeys, strict bool) error {
	reader, err := jsonl.NewReader(input)
	if err != nil {
		return err
	}
	defer reader.Close()

	seen := make(map[string]bool)
	writer, err := jsonl.NewWriter("-", false)
	if err != nil {
		return err
	}
	defer writer.Close()

	for {
		rec, err := reader.Next()
		if err != nil {
			break
		}

		var dedupKey string
		if strict {
			dedupKey = fmt.Sprintf("%p", rec)
		} else if allKeys {
			dedupKey = fmt.Sprintf("%v", rec)
		} else if key != "" {
			v := jsonl.MustGetField(rec, key)
			dedupKey = fmt.Sprintf("%v", v)
		} else {
			dedupKey = fmt.Sprintf("%v", rec)
		}

		if seen[dedupKey] {
			continue
		}
		seen[dedupKey] = true

		if err := writer.Write(rec); err != nil {
			return err
		}
	}

	return nil
}
