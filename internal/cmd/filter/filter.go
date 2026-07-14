// Package filter provides the "filter" command for filtering JSONL records.
package filter

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/EdgarOrtegaRamirez/jsonlforge/internal/pkg/jsonl"
)

// NewCommand creates the filter command.
func NewCommand() *cobra.Command {
	var (
		field string
		value string
		regex bool
	)

	cmd := &cobra.Command{
		Use:   "filter [file]",
		Short: "Filter JSONL records by field conditions",
		Long: `Filter JSONL records by field values or patterns.

Examples:
  jsonlforge filter --field 'age > 18' data.jsonl
  jsonlforge filter --field 'name == John' data.jsonl
  jsonlforge filter --field 'email !~ @gmail.com' data.jsonl --regex
  jsonlforge filter --field 'status == active' data.jsonl`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := "-"
			if len(args) > 0 && args[0] != "" {
				input = args[0]
			}
			return runFilter(input, field, value, regex)
		},
	}

	cmd.Flags().StringVar(&field, "field", "", "Field condition: 'field op value' (required)")
	cmd.Flags().StringVar(&value, "value", "", "Value to match (deprecated, use --field)")
	cmd.Flags().BoolVar(&regex, "regex", false, "Treat value as regex pattern")

	return cmd
}

type filterOp int

const (
	opEquals filterOp = iota
	opNotEquals
	opGreaterThan
	opLessThan
	opContains
	opMatches
)

func parseCondition(cond string) (field string, op filterOp, matchVal string, err error) {
	// Try operators in order of specificity
	ops := []struct {
		sym string
		op  filterOp
	}{
		{"==", opEquals},
		{"!=", opNotEquals},
		{">=", opGreaterThan},
		{"<=", opLessThan},
		{">", opGreaterThan},
		{"<", opLessThan},
		{"~", opContains},
		{"!~", opMatches},
	}

	// Split by first operator
	for _, o := range ops {
		idx := strings.Index(cond, o.sym)
		if idx > 0 {
			fld := strings.TrimSpace(cond[:idx])
			val := strings.TrimSpace(cond[idx+len(o.sym):])
			return fld, o.op, val, nil
		}
	}
	return "", 0, "", fmt.Errorf("invalid condition: %s", cond)
}

func valueMatch(rec jsonl.Record, field string, op filterOp, matchVal string) bool {
	v := jsonl.MustGetField(rec, field)
	if v == nil {
		return false
	}

	sv := fmt.Sprintf("%v", v)
	mv := matchVal

	switch op {
	case opEquals:
		return sv == mv
	case opNotEquals:
		return sv != mv
	case opGreaterThan:
		return compareNums(v, mv) > 0
	case opLessThan:
		return compareNums(v, mv) < 0
	case opContains:
		return strings.Contains(sv, mv)
	case opMatches:
		return !strings.Contains(sv, mv)
	}
	return false
}

func compareNums(v interface{}, mv string) float64 {
	fv, _ := toFloat(v)
	mvF, _ := parseFloat(mv)
	return fv - mvF
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
		f, err := parseFloat(val)
		return f, err == nil
	default:
		return 0, false
	}
}

func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

func runFilter(input, condition, value string, useRegex bool) error {
	var field string
	var op filterOp
	var matchVal string

	if condition != "" {
		f, o, mv, err := parseCondition(condition)
		if err != nil {
			return err
		}
		field, op, matchVal = f, o, mv
	} else if value != "" {
		// Legacy: --field <field> --value <value>
		field = condition
		op = opEquals
		matchVal = value
	} else {
		return fmt.Errorf("provide a condition with --field (e.g., --field 'age > 18')")
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

		if valueMatch(rec, field, op, matchVal) {
			if err := writer.Write(rec); err != nil {
				return err
			}
		}
	}

	return nil
}
