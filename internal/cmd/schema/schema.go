// Package schema provides the "schema" command for JSONL schema detection and validation.
package schema

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/EdgarOrtegaRamirez/jsonlforge/internal/pkg/jsonl"
)

// NewCommand creates the schema command.
func NewCommand() *cobra.Command {
	var (
		action string
		schema string
	)

	cmd := &cobra.Command{
		Use:   "schema [file]",
		Short: "Auto-detect or validate against JSON Schema",
		Long: `Auto-detect the schema of a JSONL file or validate against a JSON Schema.

Actions:
  detect - Auto-detect the schema (field names, types, counts)
  check  - Validate records against a JSON Schema file

Examples:
  jsonlforge schema --action detect data.jsonl
  jsonlforge schema --action check --schema schema.json data.jsonl`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := "-"
			if len(args) > 0 && args[0] != "" {
				input = args[0]
			}
			return runSchema(input, action, schema)
		},
	}

	cmd.Flags().StringVar(&action, "action", "detect", "Action: detect, check")
	cmd.Flags().StringVar(&schema, "schema", "", "Path to JSON Schema file (for check action)")

	cmd.MarkFlagRequired("action")

	return cmd
}

type fieldInfo struct {
	types    map[string]int
	count    int
}

func detectSchema(input string) error {
	reader, err := jsonl.NewReader(input)
	if err != nil {
		return err
	}
	defer reader.Close()

	fieldInfoMap := make(map[string]*fieldInfo)
	totalRecords := 0

	for {
		rec, err := reader.Next()
		if err != nil {
			break
		}
		totalRecords++

		for k, v := range rec {
			fi, ok := fieldInfoMap[k]
			if !ok {
				fi = &fieldInfo{types: make(map[string]int)}
				fieldInfoMap[k] = fi
			}
			fi.count++
			fi.types[jsonType(v)]++
		}
	}

	// Sort field names for consistent output
	keys := make([]string, 0, len(fieldInfoMap))
	for k := range fieldInfoMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	fmt.Println("JSONL Schema Detection Results")
	fmt.Println("==============================")
	fmt.Printf("Total records: %d\n", totalRecords)
	fmt.Printf("Total fields:  %d\n\n", len(fieldInfoMap))

	for _, k := range keys {
		fi := fieldInfoMap[k]
		present := fi.count
		missing := totalRecords - fi.count
		pct := float64(present) / float64(totalRecords) * 100

		// Determine primary type
		primaryType := ""
		maxCount := 0
		for t, c := range fi.types {
			if c > maxCount {
				maxCount = c
				primaryType = t
			}
		}

		fmt.Printf("  %s: type=%s present=%d/%d (%.1f%%) missing=%d",
			k, primaryType, present, totalRecords, pct, missing)

		if len(fi.types) > 1 {
			fmt.Print(" types=[")
			typeNames := make([]string, 0, len(fi.types))
			for t := range fi.types {
				typeNames = append(typeNames, t)
			}
			sort.Strings(typeNames)
			fmt.Print(strings.Join(typeNames, ", "))
			fmt.Print("]")
		}
		fmt.Println()
	}

	return nil
}

func jsonType(v interface{}) string {
	switch v.(type) {
	case string:
		return "string"
	case float64:
		return "number"
	case bool:
		return "boolean"
	case map[string]interface{}:
		return "object"
	case []interface{}:
		return "array"
	case nil:
		return "null"
	default:
		return "unknown"
	}
}

func runSchema(input, action, schemaPath string) error {
	switch action {
	case "detect":
		return detectSchema(input)
	case "check":
		if schemaPath == "" {
			return fmt.Errorf("--schema is required for check action")
		}
		return fmt.Errorf("schema validation: JSON Schema support coming soon")
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}
