// Package merge provides the "merge" command for merging multiple JSONL files by key.
package merge

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/EdgarOrtegaRamirez/jsonlforge/internal/pkg/jsonl"
)

// NewCommand creates the merge command.
func NewCommand() *cobra.Command {
	var (
		key        string
		strategy   string
		outputFile string
	)

	cmd := &cobra.Command{
		Use:   "merge [file1 file2 ...]",
		Short: "Merge multiple JSONL files by key",
		Long: `Merge multiple JSONL files by a common key field.

Merges records from multiple files, matching on the specified key field.
Records with the same key from different files are merged together.
Records present in only one file are included as-is.

Strategies:
  left    - Use fields from the first file (default)
  right   - Use fields from the last file
  combined - Merge all fields from all files

Examples:
  jsonlforge merge --key 'id' file1.jsonl file2.jsonl
  jsonlforge merge --key 'user_id' --strategy combined a.jsonl b.jsonl c.jsonl`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("at least two input files are required")
			}
			return runMerge(args, key, strategy, outputFile)
		},
	}

	cmd.Flags().StringVar(&key, "key", "", "Field to match records by (required)")
	cmd.Flags().StringVar(&strategy, "strategy", "left", "Merge strategy: left, right, combined")
	cmd.Flags().StringVar(&outputFile, "output", "-", "Output file (- for stdout)")

	cmd.MarkFlagRequired("key")

	return cmd
}

func runMerge(files []string, key, strategy, outputFile string) error {
	// Read all records, grouped by key
	type fileRecord struct {
		rec jsonl.Record
		idx int
	}
	byKey := make(map[string][]fileRecord)

	for fileIdx, file := range files {
		reader, err := jsonl.NewReader(file)
		if err != nil {
			return fmt.Errorf("opening %s: %w", file, err)
		}

		for {
			rec, err := reader.Next()
			if err != nil {
				break
			}

			kv := jsonl.MustGetField(rec, key)
			if kv == nil {
				continue
			}
			groupKey := fmt.Sprintf("%v", kv)
			byKey[groupKey] = append(byKey[groupKey], fileRecord{rec: rec, idx: fileIdx})
		}
		reader.Close()
	}

	writer, err := jsonl.NewWriter(outputFile, false)
	if err != nil {
		return err
	}
	defer writer.Close()

	for groupKey, records := range byKey {
		var merged jsonl.Record

		switch strategy {
		case "right":
			// Last file wins
			last := records[len(records)-1]
			merged = make(jsonl.Record)
			for k, v := range last.rec {
				merged[k] = v
			}
		case "combined":
			// Merge all fields
			merged = make(jsonl.Record)
			for _, fr := range records {
				for k, v := range fr.rec {
					merged[k] = v
				}
			}
		default: // left
			first := records[0]
			merged = make(jsonl.Record)
			for k, v := range first.rec {
				merged[k] = v
			}
		}

		merged[key] = groupKey
		if err := writer.Write(merged); err != nil {
			return err
		}
	}

	return nil
}
