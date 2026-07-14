// Package convert provides the "convert" command for JSONL format conversion.
package convert

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/EdgarOrtegaRamirez/jsonlforge/internal/pkg/jsonl"
)

// NewCommand creates the convert command.
func NewCommand() *cobra.Command {
	var (
		to     string
		sep    string
		indent int
	)

	cmd := &cobra.Command{
		Use:   "convert [file]",
		Short: "Convert JSONL to CSV, TSV, or pretty JSON",
		Long: `Convert JSONL files to other formats.

Supported output formats:
  csv  - Convert to CSV (comma-separated values)
  tsv  - Convert to TSV (tab-separated values)
  json - Convert to pretty-printed JSON array

Examples:
  jsonlforge convert --to csv data.jsonl > data.csv
  jsonlforge convert --to tsv data.jsonl
  jsonlforge convert --to json data.jsonl`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := "-"
			if len(args) > 0 && args[0] != "" {
				input = args[0]
			}
			return runConvert(input, to, sep, indent)
		},
	}

	cmd.Flags().StringVar(&to, "to", "csv", "Output format: csv, tsv, json")
	cmd.Flags().StringVar(&sep, "sep", ",", "Field separator (for csv)")
	cmd.Flags().IntVar(&indent, "indent", 2, "Indent level for JSON output")

	return cmd
}

func toCSVField(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return fmt.Sprintf("%v", val)
	case bool:
		return fmt.Sprintf("%v", val)
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", val)
	}
}

func runConvert(input, format, sep string, indent int) error {
	reader, err := jsonl.NewReader(input)
	if err != nil {
		return err
	}
	defer reader.Close()

	// Read all records to determine headers
	records := make([]jsonl.Record, 0)
	headers := make([]string, 0)
	headerSet := make(map[string]bool)

	for {
		rec, err := reader.Next()
		if err != nil {
			break
		}
		records = append(records, rec)
		for k := range rec {
			if !headerSet[k] {
				headerSet[k] = true
				headers = append(headers, k)
			}
		}
	}

	switch format {
	case "csv", "tsv":
		if sep == "" {
			sep = ","
		}
		if format == "tsv" {
			sep = "\t"
		}
		w := csv.NewWriter(os.Stdout)
		w.Comma = rune(sep[0])

		// Write header
		if err := w.Write(headers); err != nil {
			return err
		}

		// Write rows
		for _, rec := range records {
			row := make([]string, len(headers))
			for i, h := range headers {
				row[i] = toCSVField(rec[h])
			}
			if err := w.Write(row); err != nil {
				return err
			}
		}
		w.Flush()
		return w.Error()

	case "json":
		arr := make([]json.RawMessage, len(records))
		for i, rec := range records {
			raw, err := json.Marshal(rec)
			if err != nil {
				return err
			}
			arr[i] = raw
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", strings.Repeat(" ", indent))
		return enc.Encode(arr)
	}

	return fmt.Errorf("unsupported format: %s", format)
}
