// Package stats provides the "stats" command for computing summary statistics.
package stats

import (
	"fmt"
	"math"
	"sort"

	"github.com/spf13/cobra"

	"github.com/EdgarOrtegaRamirez/jsonlforge/internal/pkg/jsonl"
)

// NewCommand creates the stats command.
func NewCommand() *cobra.Command {
	var (
		numeric bool
		categorical bool
	)

	cmd := &cobra.Command{
		Use:   "stats [file]",
		Short: "Compute summary statistics for JSONL fields",
		Long: `Compute summary statistics for fields in a JSONL file.

Shows count, missing, type, mean, min, max, std dev for numeric fields
and top values for categorical fields.

Examples:
  jsonlforge stats data.jsonl
  jsonlforge stats --numeric data.jsonl
  jsonlforge stats --categorical data.jsonl`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := "-"
			if len(args) > 0 && args[0] != "" {
				input = args[0]
			}
			return runStats(input, numeric, categorical)
		},
	}

	cmd.Flags().BoolVar(&numeric, "numeric", false, "Show only numeric fields")
	cmd.Flags().BoolVar(&categorical, "categorical", false, "Show only categorical fields")

	return cmd
}

type numericStats struct {
	count    int
	mean     float64
	min      float64
	max      float64
	stdDev   float64
	sum      float64
	missing  int
}

func runStats(input string, numericOnly, categoricalOnly bool) error {
	reader, err := jsonl.NewReader(input)
	if err != nil {
		return err
	}
	defer reader.Close()

	type fieldStats struct {
		count     int
		missing   int
		isNumeric bool
		values    []float64
		strCounts map[string]int
		total     int
	}

	fs := make(map[string]*fieldStats)
	totalRecords := 0

	for {
		rec, err := reader.Next()
		if err != nil {
			break
		}
		totalRecords++

		for k, v := range rec {
			f, ok := fs[k]
			if !ok {
				f = &fieldStats{strCounts: make(map[string]int)}
				fs[k] = f
			}
			f.total++

			if v == nil {
				f.missing++
				continue
			}

			if num, ok := toFloat(v); ok {
				f.isNumeric = true
				f.values = append(f.values, num)
			} else {
				f.strCounts[fmt.Sprintf("%v", v)]++
			}
		}
	}

	fmt.Println("JSONL Statistics")
	fmt.Println("================")
	fmt.Printf("Total records: %d\n\n", totalRecords)

	// Sort field names
	keys := make([]string, 0, len(fs))
	for k := range fs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		f := fs[k]

		if numericOnly && !f.isNumeric {
			continue
		}
		if categoricalOnly && f.isNumeric {
			continue
		}

		fmt.Printf("Field: %s\n", k)
		fmt.Printf("  Total: %d\n", f.total)
		fmt.Printf("  Missing: %d (%.1f%%)\n", f.missing, float64(f.missing)/float64(f.total)*100)

		if f.isNumeric && len(f.values) > 0 {
			count := len(f.values)
			sum := 0.0
			minVal := math.MaxFloat64
			maxVal := -math.MaxFloat64

			for _, v := range f.values {
				sum += v
				if v < minVal {
					minVal = v
				}
				if v > maxVal {
					maxVal = v
				}
			}
			mean := sum / float64(count)

			variance := 0.0
			for _, v := range f.values {
				diff := v - mean
				variance += diff * diff
			}
			variance /= float64(count)
			stdDev := math.Sqrt(variance)

			fmt.Printf("  Type: number\n")
			fmt.Printf("  Count: %d\n", count)
			fmt.Printf("  Mean: %.2f\n", mean)
			fmt.Printf("  StdDev: %.2f\n", stdDev)
			fmt.Printf("  Min: %.2f\n", minVal)
			fmt.Printf("  Max: %.2f\n", maxVal)
		} else {
			// Show top 5 values
			typeCount := make([]struct {
				val  string
				count int
			}, 0, len(f.strCounts))
			for val, cnt := range f.strCounts {
				typeCount = append(typeCount, struct {
					val  string
					count int
				}{val: val, count: cnt})
			}
			sort.Slice(typeCount, func(i, j int) bool {
				return typeCount[i].count > typeCount[j].count
			})

			fmt.Printf("  Type: string\n")
			fmt.Printf("  Unique: %d\n", len(f.strCounts))
			fmt.Printf("  Top values:\n")
			for i, tc := range typeCount {
				if i >= 5 {
					break
				}
				pct := float64(tc.count) / float64(f.total) * 100
				fmt.Printf("    %s: %d (%.1f%%)\n", truncate(tc.val, 30), tc.count, pct)
			}
		}
		fmt.Println()
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

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
