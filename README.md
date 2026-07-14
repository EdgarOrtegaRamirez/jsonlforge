# JsonlForge — Comprehensive JSONL Processing Toolkit

**JsonlForge** is a fast, powerful CLI toolkit for processing JSONL (JSON Lines) files. Designed for data pipelines, log analysis, and shell scripting — it provides a set of composable commands for filtering, transforming, querying, validating, and analyzing JSONL data.

## Why JsonlForge?

JSONL is the de facto standard for streaming data — used by log aggregators (JSONL, cloud providers, and data engineers. But existing tools fall short:

- **jq** is powerful but has a steep learning curve
- **dataquery** uses SQL syntax which is overkill for simple operations
- **csvforge** handles CSV/TSV but not JSONL specifically
- No dedicated tool for JSONL schema detection, statistics, and aggregation

JsonlForge fills this gap with a focused, Unix-pipe-friendly toolkit.

## Features

- **12 commands**: validate, query, filter, transform, flatten, aggregate, schema, sort, dedup, merge, convert, stats
- **Streaming**: Processes files line-by-line — no need to load everything into memory
- **Shell-friendly**: All commands read from stdin or files, output to stdout
- **Fast**: Written in Go for maximum performance
- **Zero dependencies**: Single binary, no runtime dependencies

## Installation

### From Source

```bash
go install github.com/EdgarOrtegaRamirez/jsonlforge/cmd/jsonlforge@latest
```

### From Binary

Download the latest release from [GitHub Releases](https://github.com/EdgarOrtegaRamirez/jsonlforge/releases).

## Usage

### Quick Start

```bash
# Validate a JSONL file
jsonlforge validate data.jsonl

# View summary statistics
jsonlforge stats data.jsonl

# Query specific fields
jsonlforge query --fields 'name,age' data.jsonl

# Filter records
jsonlforge filter --field 'age > 18' data.jsonl

# Aggregate by category
jsonlforge aggregate --key 'city' --agg 'count:*,avg:age' data.jsonl
```

### Command Reference

#### validate
Validate JSONL file structure, checking for valid JSON objects.

```bash
jsonlforge validate data.jsonl
jsonlforge validate --strict data.jsonl
```

#### query
Query JSONL files to select specific fields.

```bash
jsonlforge query --fields 'name,age,city' data.jsonl
jsonlforge query --all --limit 10 data.jsonl
```

#### filter
Filter records by field conditions.

```bash
jsonlforge filter --field 'age > 18' data.jsonl
jsonlforge filter --field 'status == active' data.jsonl
jsonlforge filter --field 'name ~ John' data.jsonl
```

#### transform
Transform records with field mapping and renaming.

```bash
jsonlforge transform --fields 'id,name,email' data.jsonl
jsonlforge transform --rename 'first_name:name,last_name:email' data.jsonl
```

#### flatten
Flatten nested JSON objects into flat key-value pairs.

```bash
jsonlforge flatten data.jsonl
jsonlforge flatten --sep '_' data.jsonl
```

#### aggregate
Aggregate records by a key field with computed aggregations.

```bash
jsonlforge aggregate --key 'city' --agg 'count:*,avg:age' data.jsonl
jsonlforge aggregate --key 'product' --agg 'sum:price,min:price,max:price' data.jsonl
```

#### schema
Auto-detect the schema of a JSONL file.

```bash
jsonlforge schema --action detect data.jsonl
```

#### sort
Sort records by field value.

```bash
jsonlforge sort --field 'name' data.jsonl
jsonlforge sort --field 'age' --desc data.jsonl
```

#### dedup
Remove duplicate records.

```bash
jsonlforge dedup data.jsonl
jsonlforge dedup --key 'id' data.jsonl
```

#### merge
Merge multiple JSONL files by key.

```bash
jsonlforge merge --key 'id' file1.jsonl file2.jsonl
jsonlforge merge --key 'user_id' --strategy combined a.jsonl b.jsonl
```

#### convert
Convert JSONL to CSV, TSV, or pretty JSON.

```bash
jsonlforge convert --to csv data.jsonl > data.csv
jsonlforge convert --to tsv data.jsonl
jsonlforge convert --to json data.jsonl
```

#### stats
Compute summary statistics for fields.

```bash
jsonlforge stats data.jsonl
jsonlforge stats --numeric data.jsonl
jsonlforge stats --categorical data.jsonl
```

## Piping Examples

```bash
# Validate then filter then aggregate
jsonlforge validate data.jsonl && \
jsonlforge filter --field 'age > 25' data.jsonl | \
jsonlforge aggregate --key 'city' --agg 'count:*,avg:age' -

# Convert to CSV
jsonlforge query --fields 'name,age,city' data.jsonl | \
jsonlforge convert --to csv -

# Flatten and aggregate
jsonlforge flatten data.jsonl | \
jsonlforge aggregate --key 'address_city' --agg 'count:*' -
```

## Architecture

```
jsonlforge/
├── cmd/jsonlforge/main.go        # Entry point
├── internal/
│   ├── cmd/                      # Command implementations
│   │   ├── root.go               # Root command and CLI structure
│   │   ├── validate/             # validate command
│   │   ├── query/                # query command
│   │   ├── filter/               # filter command
│   │   ├── transform/            # transform command
│   │   ├── flatten/              # flatten command
│   │   ├── aggregate/            # aggregate command
│   │   ├── schema/               # schema command
│   │   ├── sort/                 # sort command
│   │   ├── dedup/                # dedup command
│   │   ├── merge/                # merge command
│   │   ├── convert/              # convert command
│   │   └── stats/                # stats command
│   └── pkg/jsonl/                # Core JSONL reading/writing library
├── LICENSE                       # MIT License
├── README.md                     # This file
└── .github/workflows/go.yml      # CI/CD pipeline
```

## Comparison with Similar Tools

| Feature | JsonlForge | jq | dataquery | csvforge |
|---------|-----------|----|-----------|----------|
| JSONL-specific | ✅ | ❌ | ❌ | ❌ |
| Schema detection | ✅ | ❌ | ❌ | ❌ |
| Statistics | ✅ | ❌ | ❌ | ❌ |
| Aggregation | ✅ | ✅ (complex) | ✅ (SQL) | ❌ |
| Filter by condition | ✅ | ✅ (complex) | ✅ (SQL) | ❌ |
| Easy learning curve | ✅ | ❌ | ❌ | ✅ |
| Single binary | ✅ | ✅ | ❌ | ✅ |

## License

[MIT License](LICENSE)
