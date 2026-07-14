// Package sort provides the "sort" command for sorting JSONL records.
package sort

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/EdgarOrtegaRamirez/jsonlforge/internal/pkg/jsonl"
)

// NewCommand creates the sort command.
func NewCommand() *cobra.Command {
	var (
		field string
		desc  bool
	)

	cmd := &cobra.Command{
		Use:   "sort [file]",
		Short: "Sort JSONL records by field value",
		Long: `Sort JSONL records by a field value.

Sorting supports string, numeric, and date comparisons.
Use --desc for descending order.

Examples:
  jsonlforge sort --field 'name' data.jsonl
  jsonlforge sort --field 'age' --desc data.jsonl
  jsonlforge sort --field 'created_at' data.jsonl`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := "-"
			if len(args) > 0 && args[0] != "" {
				input = args[0]
			}
			return runSort(input, field, desc)
		},
	}

	cmd.Flags().StringVar(&field, "field", "", "Field to sort by (required)")
	cmd.Flags().BoolVar(&desc, "desc", false, "Sort in descending order")

	cmd.MarkFlagRequired("field")

	return cmd
}

func runSort(input, field string, desc bool) error {
	reader, err := jsonl.NewReader(input)
	if err != nil {
		return err
	}
	defer reader.Close()

	var records []jsonl.Record
	for {
		rec, err := reader.Next()
		if err != nil {
			break
		}
		records = append(records, rec)
	}

	sort.SliceStable(records, func(i, j int) bool {
		a := jsonl.MustGetField(records[i], field)
		b := jsonl.MustGetField(records[j], field)

		ai, aok := toFloat(a)
		bi, bok := toFloat(b)

		if aok && bok {
			if desc {
				return ai > bi
			}
			return ai < bi
		}

		sa := fmt.Sprintf("%v", a)
		sb := fmt.Sprintf("%v", b)
		if desc {
			return sa > sb
		}
		return sa < sb
	})

	writer, err := jsonl.NewWriter("-", false)
	if err != nil {
		return err
	}
	defer writer.Close()

	for _, rec := range records {
		if err := writer.Write(rec); err != nil {
			return err
		}
	}

	return nil
}

func toFloat(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case string:
		var f float64
		_, err := fmt.Sscanf(val, "%f", &f)
		return f, err == nil
	default:
		return 0, false
	}
}
