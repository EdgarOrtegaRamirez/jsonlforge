// Package flatten provides the "flatten" command for flattening nested JSON objects.
package flatten

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/EdgarOrtegaRamirez/jsonlforge/internal/pkg/jsonl"
)

// NewCommand creates the flatten command.
func NewCommand() *cobra.Command {
	var (
		sep    string
		prefix string
	)

	cmd := &cobra.Command{
		Use:   "flatten [file]",
		Short: "Flatten nested JSON objects into flat key-value pairs",
		Long: `Flatten nested JSONL objects into flat key-value pairs.

Nested objects are flattened using a configurable separator (default: dot).
Arrays are indexed numerically.

Examples:
  jsonlforge flatten data.jsonl
  jsonlforge flatten --sep '_' data.jsonl
  jsonlforge flatten --prefix 'root.' data.jsonl`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := "-"
			if len(args) > 0 && args[0] != "" {
				input = args[0]
			}
			return runFlatten(input, sep, prefix)
		},
	}

	cmd.Flags().StringVar(&sep, "sep", ".", "Separator for nested keys")
	cmd.Flags().StringVar(&prefix, "prefix", "", "Prefix for all keys")

	return cmd
}

func flattenValue(obj interface{}, prefix string, result map[string]interface{}, sep string) {
	if obj == nil {
		if prefix != "" {
			result[prefix] = nil
		}
		return
	}

	switch v := obj.(type) {
	case map[string]interface{}:
		for k, val := range v {
			var newKey string
			if prefix != "" {
				newKey = prefix + sep + k
			} else {
				newKey = k
			}
			flattenValue(val, newKey, result, sep)
		}
	case []interface{}:
		for i, val := range v {
			var newKey string
			if prefix != "" {
				newKey = prefix + sep + fmt.Sprintf("%d", i)
			} else {
				newKey = fmt.Sprintf("%d", i)
			}
			flattenValue(val, newKey, result, sep)
		}
	default:
		if prefix != "" {
			result[prefix] = v
		} else {
			result["value"] = v
		}
	}
}

func runFlatten(input, sep, prefix string) error {
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

		flattened := make(map[string]interface{})
		flattenValue(rec, prefix, flattened, sep)

		if err := writer.Write(flattened); err != nil {
			return err
		}
	}

	return nil
}
