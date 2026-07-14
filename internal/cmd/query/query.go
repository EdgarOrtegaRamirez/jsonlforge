// Package query provides the "query" command for querying JSONL files.
package query

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/EdgarOrtegaRamirez/jsonlforge/internal/pkg/jsonl"
)

// NewCommand creates the query command.
func NewCommand() *cobra.Command {
	var (
		fields    string
		selectAll bool
		limit     int
	)

	cmd := &cobra.Command{
		Use:   "query [file]",
		Short: "Query JSONL files with jq-like expression syntax",
		Long: `Query JSONL files to select specific fields or expressions.

Select fields with --fields (comma-separated). Select all with --all.
Limit output with --limit.

Examples:
  jsonlforge query --fields 'name,age' data.jsonl
  jsonlforge query --all --limit 10 data.jsonl
  jsonlforge query --fields 'name' data.jsonl | head`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := "-"
			if len(args) > 0 && args[0] != "" {
				input = args[0]
			}
			return runQuery(input, fields, selectAll, limit)
		},
	}

	cmd.Flags().StringVar(&fields, "fields", "", "Comma-separated list of fields to select")
	cmd.Flags().BoolVar(&selectAll, "all", false, "Select all fields")
	cmd.Flags().IntVar(&limit, "limit", 0, "Limit number of output records")

	return cmd
}

func runQuery(input, fields string, selectAll bool, limit int) error {
	reader, err := jsonl.NewReader(input)
	if err != nil {
		return err
	}
	defer reader.Close()

	var wantedFields []string
	if fields != "" {
		for _, f := range strings.Split(fields, ",") {
			f = strings.TrimSpace(f)
			if f != "" {
				wantedFields = append(wantedFields, f)
			}
		}
	}

	writer, err := jsonl.NewWriter("-", false)
	if err != nil {
		return err
	}
	defer writer.Close()

	count := 0
	for {
		rec, err := reader.Next()
		if err != nil {
			break
		}

		if limit > 0 && count >= limit {
			break
		}

		if selectAll {
			if err := writer.Write(rec); err != nil {
				return err
			}
		} else if len(wantedFields) > 0 {
			result := make(jsonl.Record)
			for _, f := range wantedFields {
				if v, ok := rec[f]; ok {
					result[f] = v
				}
			}
			if len(result) > 0 {
				if err := writer.Write(result); err != nil {
					return err
				}
			}
		} else {
			// No fields specified, output first field value
			if len(rec) > 0 {
				for k, v := range rec {
					fmt.Printf("%s: %s\n", k, toJSON(v))
					break
				}
			}
		}
		count++
	}

	return nil
}

func toJSON(v interface{}) string {
	bytes, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(bytes)
}
