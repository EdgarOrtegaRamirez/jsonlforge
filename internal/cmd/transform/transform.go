// Package transform provides the "transform" command for transforming JSONL records.
package transform

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/EdgarOrtegaRamirez/jsonlforge/internal/pkg/jsonl"
)

// NewCommand creates the transform command.
func NewCommand() *cobra.Command {
	var (
		fields string
		keep   string
		rename string
	)

	cmd := &cobra.Command{
		Use:   "transform [file]",
		Short: "Transform JSONL records with field mapping, renaming, and deletion",
		Long: `Transform JSONL records with field operations.

Operations:
  --fields   Comma-separated list of fields to keep (drops others)
  --keep     Alias for --fields
  --rename   Rename fields: 'old:new,old2:new2'

Examples:
  jsonlforge transform --fields 'name,email,age' data.jsonl
  jsonlforge transform --rename 'first_name:name,last_name:email' data.jsonl
  jsonlforge transform --fields 'id,name' --rename 'name:title' data.jsonl`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := "-"
			if len(args) > 0 && args[0] != "" {
				input = args[0]
			}
			return runTransform(input, fields, keep, rename)
		},
	}

	cmd.Flags().StringVar(&fields, "fields", "", "Comma-separated list of fields to keep")
	cmd.Flags().StringVar(&keep, "keep", "", "Alias for --fields")
	cmd.Flags().StringVar(&rename, "rename", "", "Rename fields: 'old:new,old2:new2'")

	return cmd
}

func runTransform(input, fields, keep, rename string) error {
	// Merge --fields and --keep
	if keep != "" && fields == "" {
		fields = keep
	}

	// Parse rename map
	renameMap := make(map[string]string)
	if rename != "" {
		for _, pair := range strings.Split(rename, ",") {
			parts := strings.SplitN(pair, ":", 2)
			if len(parts) == 2 {
				old := strings.TrimSpace(parts[0])
				new := strings.TrimSpace(parts[1])
				renameMap[old] = new
			}
		}
	}

	// Parse fields to keep
	var keepFields []string
	if fields != "" {
		for _, f := range strings.Split(fields, ",") {
			f = strings.TrimSpace(f)
			if f != "" {
				keepFields = append(keepFields, f)
			}
		}
	}

	reader, err := jsonl.NewReader(input)
	if err != nil {
		return err
	}
	defer reader.Close()

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

		// Apply field selection
		var result jsonl.Record
		if len(keepFields) > 0 {
			result = make(jsonl.Record)
			for _, f := range keepFields {
				if v, ok := rec[f]; ok {
					result[f] = v
				}
			}
		} else {
			result = make(jsonl.Record)
			for k, v := range rec {
				result[k] = v
			}
		}

		// Apply renames
		if len(renameMap) > 0 {
			newResult := make(jsonl.Record)
			for k, v := range result {
				if newK, ok := renameMap[k]; ok {
					newResult[newK] = v
				} else {
					newResult[k] = v
				}
			}
			result = newResult
		}

		if err := writer.Write(result); err != nil {
			return err
		}
	}

	return nil
}
