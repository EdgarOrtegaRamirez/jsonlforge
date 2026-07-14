// Package aggregate provides the "aggregate" command for aggregating JSONL records.
package aggregate

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/EdgarOrtegaRamirez/jsonlforge/internal/pkg/jsonl"
)

// NewCommand creates the aggregate command.
func NewCommand() *cobra.Command {
	var (
		key     string
		aggs    string
		output  string
	)

	cmd := &cobra.Command{
		Use:   "aggregate [file]",
		Short: "Aggregate JSONL records by key with count, sum, avg, min, max",
		Long: `Aggregate JSONL records by a key field with computed aggregations.

Supports these aggregation functions:
  count:*  - Count of records in each group
  sum:FIELD - Sum of a numeric field
  avg:FIELD - Average of a numeric field
  min:FIELD - Minimum value
  max:FIELD - Maximum value
  first:FIELD - First value seen
  last:FIELD  - Last value seen
  distinct:FIELD - Count of distinct values

Examples:
  jsonlforge aggregate --key 'city' --agg 'count:*,sum:salary' sales.jsonl
  jsonlforge aggregate --key 'product' --agg 'count:*,avg:price,min:price,max:price' orders.jsonl
  jsonlforge aggregate --key 'category' --agg 'distinct:product' catalog.jsonl`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := "-"
			if len(args) > 0 && args[0] != "" {
				input = args[0]
			}
			return runAggregate(input, key, aggs, output)
		},
	}

	cmd.Flags().StringVar(&key, "key", "", "Field to group by (required)")
	cmd.Flags().StringVar(&aggs, "agg", "", "Aggregation expressions, comma-separated (required)")
	cmd.Flags().StringVar(&output, "output", "-", "Output file (- for stdout)")

	cmd.MarkFlagRequired("key")
	cmd.MarkFlagRequired("agg")

	return cmd
}

type aggFunc struct {
	name string
	field string
}

type aggResult struct {
	count int
	sum   float64
	min   float64
	max   float64
	first interface{}
	last  interface{}
	dists map[string]int
}

func parseAggExpr(s string) aggFunc {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) == 2 {
		return aggFunc{name: strings.TrimSpace(parts[0]), field: strings.TrimSpace(parts[1])}
	}
	return aggFunc{name: strings.TrimSpace(s), field: ""}
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
		f, err := strconv.ParseFloat(val, 64)
		return f, err == nil
	case jsonl.Record:
		return 0, false
	default:
		return 0, false
	}
}

func runAggregate(input, key, aggs string, output string) error {
	reader, err := jsonl.NewReader(input)
	if err != nil {
		return err
	}
	defer reader.Close()

	parsedAggs := make([]aggFunc, 0)
	for _, a := range strings.Split(aggs, ",") {
		a = strings.TrimSpace(a)
		if a != "" {
			parsedAggs = append(parsedAggs, parseAggExpr(a))
		}
	}

	groups := make(map[string]*aggResult)

	for {
		rec, err := reader.Next()
		if err != nil {
			break
		}

		kv := jsonl.MustGetField(rec, key)
		groupKey := fmt.Sprintf("%v", kv)

		grp, ok := groups[groupKey]
		if !ok {
			grp = &aggResult{
				min:   math.MaxFloat64,
				max:   -math.MaxFloat64,
				dists: make(map[string]int),
			}
			groups[groupKey] = grp
		}
		grp.count++

		for _, agg := range parsedAggs {
			switch agg.name {
			case "count":
				// count is tracked by grp.count
			case "sum", "avg", "min", "max":
				if agg.field == "*" {
					continue
				}
				val, ok := toFloat(jsonl.MustGetField(rec, agg.field))
				if ok {
					grp.sum += val
					if val < grp.min {
						grp.min = val
					}
					if val > grp.max {
						grp.max = val
					}
				}
			case "first":
				if grp.count == 1 {
					grp.first = jsonl.MustGetField(rec, agg.field)
				}
			case "last":
				grp.last = jsonl.MustGetField(rec, agg.field)
			case "distinct":
				v := jsonl.MustGetField(rec, agg.field)
				grp.dists[fmt.Sprintf("%v", v)]++
			}
		}
	}

	writer, err := jsonl.NewWriter(output, false)
	if err != nil {
		return err
	}
	defer writer.Close()

	for groupKey, grp := range groups {
		result := jsonl.Record{key: groupKey}
		result["count"] = grp.count

		for _, agg := range parsedAggs {
			var label string
			if agg.field != "" && agg.field != "*" {
				label = agg.field
			} else {
				label = "*"
			}
			name := fmt.Sprintf("%s_%s", agg.name, label)

			switch agg.name {
			case "count":
				// already set
			case "sum":
				result[name] = round(grp.sum)
			case "avg":
				if grp.count > 0 {
					result[name] = round(grp.sum / float64(grp.count))
				}
			case "min":
				if grp.min != math.MaxFloat64 {
					result[name] = round(grp.min)
				}
			case "max":
				if grp.max != -math.MaxFloat64 {
					result[name] = round(grp.max)
				}
			case "first":
				result[name] = grp.first
			case "last":
				result[name] = grp.last
			case "distinct":
				result[name] = len(grp.dists)
			}
		}

		if err := writer.Write(result); err != nil {
			return err
		}
	}

	return nil
}

func round(v float64) float64 {
	return math.Round(v*1e6) / 1e6
}
