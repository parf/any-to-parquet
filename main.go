package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/parf/homebase-go-lib/fileiterator"
)

var (
	sqlFlag    = flag.String("sql", "", "SQL query to execute")
	tableFlag  = flag.String("table", "", "Table name (generates SELECT * FROM table)")
	driverFlag = flag.String("driver", "mysql", "Database driver: mysql or postgre")
	dsnFlag    = flag.String("dsn", "", "Database connection string")
)

func main() {
	flag.Parse()

	// Detect SQL mode
	isSQLMode := *sqlFlag != "" || *tableFlag != "" || *dsnFlag != "" || *driverFlag != "mysql"

	if isSQLMode {
		handleSQLMode()
	} else {
		handleFileMode()
	}
}

func handleFileMode() {
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "any2parquet - Convert any format to Parquet (RECOMMENDED) 🏆\n")
		fmt.Fprintf(os.Stderr, "=============================================================\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  File mode:  %s <input-file> [output-file|-]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  SQL mode:   %s --dsn=\"user:pass@host\" --sql=\"SELECT * FROM table\" [output-file|-]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "              Use '-' for stdout, omit for auto-generated filename\n\n")

		fmt.Fprintf(os.Stderr, "What is Parquet (.parquet)?\n")
		fmt.Fprintf(os.Stderr, "  Parquet is a columnar storage format optimized for analytics.\n")
		fmt.Fprintf(os.Stderr, "  Winner in benchmarks: Fastest overall (0.61s), excellent compression (44MB).\n")
		fmt.Fprintf(os.Stderr, "  Best for: Everything - APIs, analytics, data warehouses.\n")
		fmt.Fprintf(os.Stderr, "  Compatible with: Spark, DuckDB, Pandas, Arrow, all major data tools.\n\n")

		fmt.Fprintf(os.Stderr, "=== SQL DATABASE SUPPORT ===\n\n")
		fmt.Fprintf(os.Stderr, "Export directly from MySQL or PostgreSQL to Parquet format:\n\n")

		fmt.Fprintf(os.Stderr, "Flags:\n")
		fmt.Fprintf(os.Stderr, "  --sql=\"SELECT * FROM table\"  SQL query to execute\n")
		fmt.Fprintf(os.Stderr, "  --table=\"schema.table\"       Table name (alternative to --sql)\n")
		fmt.Fprintf(os.Stderr, "  --driver=mysql               Database driver: mysql or postgre (default: mysql)\n")
		fmt.Fprintf(os.Stderr, "  --dsn=\"connection-string\"    Database connection string\n\n")

		fmt.Fprintf(os.Stderr, "DSN Format:\n")
		fmt.Fprintf(os.Stderr, "  MySQL:      user:password@tcp(host:3306)/database\n")
		fmt.Fprintf(os.Stderr, "  PostgreSQL: host=localhost port=5432 user=myuser password=mypass sslmode=disable\n")
		fmt.Fprintf(os.Stderr, "  Simplified: user:password@host (auto-expanded)\n\n")

		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  %s --dsn=\"root:pass@localhost\" --sql=\"SELECT * FROM users\" users.parquet\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --dsn=\"root:pass@localhost\" --table=\"mydb.users\" - | ./any2jsonl - | jq\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --driver=postgre --dsn=\"user:pass@pghost\" --table=\"public.orders\" orders.parquet\n\n", os.Args[0])

		fmt.Fprintf(os.Stderr, "=== FILE CONVERSION ===\n\n")

		fmt.Fprintf(os.Stderr, "Supported input formats (recognized extension → format):\n")
		fmt.Fprintf(os.Stderr, "  - JSONL: JSON Lines, one JSON object per line → .jsonl\n")
		fmt.Fprintf(os.Stderr, "  - CSV: Comma-separated values with header row → .csv, .tsv, .psv\n")
		fmt.Fprintf(os.Stderr, "  - MsgPack: Binary serialization format → .msgpack\n\n")

		fmt.Fprintf(os.Stderr, "Input compression formats (recognized extension → format):\n")
		fmt.Fprintf(os.Stderr, "  .gz  → Gzip (standard compression, widely supported, slow)\n")
		fmt.Fprintf(os.Stderr, "  .zst → Zstandard (RECOMMENDED: best balance of speed & compression)\n")
		fmt.Fprintf(os.Stderr, "  .lz4 → LZ4 (fastest compression, moderate compression ratio)\n")
		fmt.Fprintf(os.Stderr, "  .br  → Brotli (best compression, but very slow)\n")
		fmt.Fprintf(os.Stderr, "  .xz  → XZ/LZMA (excellent compression, extremely slow - avoid)\n\n")

		fmt.Fprintf(os.Stderr, "⚠️  IMPORTANT: Parquet already has built-in Snappy compression!\n")
		fmt.Fprintf(os.Stderr, "   Additional compression (.parquet.gz/.zst/.lz4) is usually NOT needed.\n")
		fmt.Fprintf(os.Stderr, "   It only gives ~10-15%% smaller files but slower access.\n\n")

		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  %s data.jsonl.gz                      → data.parquet\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s data.csv -                         → stdout\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s data.csv.zst output.pq             → output.pq\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s data.jsonl data.parquet.zst        → data.parquet.zst (with Zstd)\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s data.msgpack data.parquet.lz4      → data.parquet.lz4 (with LZ4)\n\n", os.Args[0])

		fmt.Fprintf(os.Stderr, "Output compression (recognized extension → format):\n")
		fmt.Fprintf(os.Stderr, "  .parquet     → Parquet with built-in Snappy compression\n")
		fmt.Fprintf(os.Stderr, "  .parquet.zst → Parquet + Zstandard (~15%% smaller, slower access)\n")
		fmt.Fprintf(os.Stderr, "  .parquet.lz4 → Parquet + LZ4 (~10%% smaller, slower access)\n")
		fmt.Fprintf(os.Stderr, "  .parquet.gz  → Parquet + Gzip (~10-15%% smaller, slower access)\n")
		fmt.Fprintf(os.Stderr, "  \n")
		fmt.Fprintf(os.Stderr, "  ⚠️  Additional compression is usually NOT needed!\n")
		fmt.Fprintf(os.Stderr, "  Parquet already has built-in Snappy compression.\n\n")

		fmt.Fprintf(os.Stderr, "Performance (1M records):\n")
		fmt.Fprintf(os.Stderr, "  Read:  0.15s (4x faster than MsgPack, 13x faster than JSONL)\n")
		fmt.Fprintf(os.Stderr, "  Write: 0.46s (fastest binary format)\n")
		fmt.Fprintf(os.Stderr, "  Size:  44MB (72%% smaller than plain text)\n\n")

		fmt.Fprintf(os.Stderr, "Schema Support:\n")
		fmt.Fprintf(os.Stderr, "  Automatically infers schema from your data - supports ANY structure!\n")
		fmt.Fprintf(os.Stderr, "  Supported types: int64, float64, string, bool\n\n")

		fmt.Fprintf(os.Stderr, "Full Benchmark Results:\n")
		fmt.Fprintf(os.Stderr, "  https://github.com/parf/homebase-go-lib/blob/main/benchmarks/serialization-benchmark-result.md\n\n")

		fmt.Fprintf(os.Stderr, "Quick inspection with jq:\n")
		fmt.Fprintf(os.Stderr, "  ./any2jsonl data.parquet | head -5 | jq\n\n")

		fmt.Fprintf(os.Stderr, "See also: ./any2jsonl (convert to human-readable JSONL format)\n")
		os.Exit(1)
	}

	inputFile := flag.Arg(0)
	outputFile := flag.Arg(1)

	if outputFile == "" {
		outputFile = strings.TrimSuffix(inputFile, filepath.Ext(inputFile))
		// Handle double extensions like .jsonl.gz
		for _, ext := range []string{".gz", ".zst", ".lz4", ".br", ".xz"} {
			outputFile = strings.TrimSuffix(outputFile, ext)
		}
		for _, ext := range []string{".jsonl", ".csv", ".msgpack"} {
			outputFile = strings.TrimSuffix(outputFile, ext)
		}
		outputFile += ".parquet"
	}

	// Read all records from input (supports ANY schema)
	records, err := fileiterator.ReadInput(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Read %d records\n", len(records))

	// If output is "-", write to stdout. Otherwise, write to file.
	if outputFile == "-" {
		// Write to stdout (using temp file since Parquet needs seekable writer)
		tmpFile := "/tmp/any2parquet-" + fmt.Sprintf("%d", os.Getpid()) + ".parquet"
		if err := fileiterator.WriteParquetAny(tmpFile, records); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing Parquet: %v\n", err)
			os.Exit(1)
		}
		defer os.Remove(tmpFile)

		// Copy to stdout
		data, err := os.ReadFile(tmpFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading temp file: %v\n", err)
			os.Exit(1)
		}
		if _, err := os.Stdout.Write(data); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to stdout: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Fprintf(os.Stderr, "Converting %s -> %s\n", inputFile, outputFile)
		// Write to Parquet file (compression auto-detected from filename)
		if err := fileiterator.WriteParquetAny(outputFile, records); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing Parquet: %v\n", err)
			os.Exit(1)
		}
		stat, _ := os.Stat(outputFile)
		fmt.Fprintf(os.Stderr, "Written %s (%d bytes, %.2f MB)\n", outputFile, stat.Size(), float64(stat.Size())/1024/1024)
	}
}

func handleSQLMode() {
	// Validate required flags
	if *dsnFlag == "" {
		fmt.Fprintf(os.Stderr, "Error: --dsn is required for SQL queries\n")
		os.Exit(1)
	}

	if *sqlFlag == "" && *tableFlag == "" {
		fmt.Fprintf(os.Stderr, "Error: Either --sql or --table is required\n")
		os.Exit(1)
	}

	if *sqlFlag != "" && *tableFlag != "" {
		fmt.Fprintf(os.Stderr, "Error: Cannot use both --sql and --table\n")
		os.Exit(1)
	}

	// Generate SQL query from table if needed
	sqlQuery := *sqlFlag
	if sqlQuery == "" && *tableFlag != "" {
		sqlQuery = fmt.Sprintf("SELECT * FROM %s", *tableFlag)
	}

	// Get optional output file from positional argument
	outputFile := ""
	if flag.NArg() > 0 {
		outputFile = flag.Arg(0)
	}

	// Normalize DSN
	dsn := normalizeDSN(*driverFlag, *dsnFlag)

	fmt.Fprintf(os.Stderr, "Executing SQL query: %s\n", sqlQuery)

	// Read from SQL
	records, err := fileiterator.ReadSQLInput(*driverFlag, dsn, sqlQuery)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing SQL query: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Read %d records\n", len(records))

	// If output file is "-" or empty, write to stdout. Otherwise, write to file.
	if outputFile == "" || outputFile == "-" {
		// Write to stdout (using temp file since Parquet needs seekable writer)
		tmpFile := "/tmp/any2parquet-" + fmt.Sprintf("%d", os.Getpid()) + ".parquet"
		if err := fileiterator.WriteParquetAny(tmpFile, records); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing Parquet: %v\n", err)
			os.Exit(1)
		}
		defer os.Remove(tmpFile)

		// Copy to stdout
		data, err := os.ReadFile(tmpFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading temp file: %v\n", err)
			os.Exit(1)
		}
		if _, err := os.Stdout.Write(data); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to stdout: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Write to file
		if err := fileiterator.WriteParquetAny(outputFile, records); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing Parquet: %v\n", err)
			os.Exit(1)
		}
		stat, _ := os.Stat(outputFile)
		fmt.Fprintf(os.Stderr, "Written %s (%d bytes, %.2f MB)\n", outputFile, stat.Size(), float64(stat.Size())/1024/1024)
	}
}

func normalizeDSN(driver, dsn string) string {
	// If DSN contains "/" or "=" or "sslmode=", it's already in proper format
	if strings.Contains(dsn, "/") || strings.Contains(dsn, "=") || strings.Contains(dsn, "sslmode=") {
		return dsn
	}

	// Simplified format: "user:password@host"
	if !strings.Contains(dsn, "@") {
		return dsn // Return as-is if not in simplified format
	}

	parts := strings.Split(dsn, "@")
	if len(parts) != 2 {
		return dsn
	}

	userPass := parts[0]
	host := parts[1]

	switch driver {
	case "mysql":
		// MySQL: user:password@tcp(host:3306)/
		if !strings.Contains(host, ":") {
			host = host + ":3306"
		}
		return fmt.Sprintf("%s@tcp(%s)/", userPass, host)

	case "postgre", "postgres", "postgresql":
		// PostgreSQL: host=localhost port=5432 user=myuser password=mypass sslmode=disable
		userParts := strings.Split(userPass, ":")
		user := userParts[0]
		pass := ""
		if len(userParts) > 1 {
			pass = userParts[1]
		}

		hostPort := strings.Split(host, ":")
		hostName := hostPort[0]
		port := "5432"
		if len(hostPort) > 1 {
			port = hostPort[1]
		}

		return fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=disable",
			hostName, port, user, pass)
	}

	return dsn
}
