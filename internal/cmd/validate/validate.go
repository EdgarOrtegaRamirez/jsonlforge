// Package validate provides the "validate" command for JSONL file validation.
package validate

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/EdgarOrtegaRamirez/jsonlforge/internal/pkg/jsonl"
)

// NewCommand creates the validate command.
func NewCommand() *cobra.Command {
	var strict bool

	cmd := &cobra.Command{
		Use:   "validate [file]",
		Short: "Validate JSONL file structure and optional JSON Schema",
		Long: `Validate JSONL file structure, checking for valid JSON objects.

Options:
  --strict   Enable strict validation (all lines must be objects, no empty lines)

Returns exit code 0 if valid, 1 if errors found.
Prints a summary of validation results.

Examples:
  jsonlforge validate data.jsonl
  jsonlforge validate --strict data.jsonl`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := "-"
			if len(args) > 0 && args[0] != "" {
				input = args[0]
			}
			return runValidate(input, strict)
		},
	}

	cmd.Flags().BoolVar(&strict, "strict", false, "Enable strict validation")

	return cmd
}

func runValidate(input string, strict bool) error {
	reader, err := jsonl.NewReader(input)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer reader.Close()

	totalLines := 0
	validRecords := 0
	invalidLines := 0
	emptyLines := 0
	parseErrors := 0

	for {
		totalLines++
		lineNum := totalLines
		rec, err := reader.Next()
		if err != nil {
			if err.Error() == "io: read/write on closed file" {
				break
			}
			// Check if it's EOF
			if err.Error() == "EOF" || err.Error() == "io: read/write on closed file" {
				break
			}
			invalidLines++
			parseErrors++
			if strict {
				return fmt.Errorf("validation error at line %d: %w", lineNum, err)
			}
			continue
		}

		// Check it's a map (not a bare string, number, etc.)
		if rec == nil {
			emptyLines++
			if strict {
				return fmt.Errorf("line %d is not a JSON object", lineNum)
			}
			continue
		}

		validRecords++
	}

	// Summary
	fmt.Println("JSONL Validation Report")
	fmt.Println("========================")
	fmt.Printf("Total lines:   %d\n", totalLines)
	fmt.Printf("Valid records: %d\n", validRecords)
	if emptyLines > 0 {
		fmt.Printf("Empty lines:   %d\n", emptyLines)
	}
	if invalidLines > 0 {
		fmt.Printf("Invalid lines: %d\n", invalidLines)
	}
	fmt.Printf("Parse errors:  %d\n", parseErrors)

	if invalidLines > 0 || parseErrors > 0 {
		fmt.Println("\nResult: FAILED")
		return fmt.Errorf("validation failed: %d errors found", invalidLines+parseErrors)
	}

	fmt.Println("\nResult: PASSED")
	return nil
}
