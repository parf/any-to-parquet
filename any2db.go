package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/parf/homebase-go-lib/fileiterator"
	hbsql "github.com/parf/homebase-go-lib/sql"
)

var (
	sqlFlag    = flag.String("sql", "", "Source SQL query to execute")
	tableFlag  = flag.String("table", "", "Source table name")
	driverFlag = flag.String("driver", "mysql", "Database driver: mysql or postgre")
	dsnFlag    = flag.String("dsn", "", "Destination database connection string")
	batchFlag  = flag.Int("batch", 1000, "Batch size for inserts")
)

func main() {
	flag.Parse()

	if flag.NArg() < 1 {
		showHelp()
		os.Exit(1)
	}

	// Check if DSN is provided
	if *dsnFlag == "" {
		fmt.Fprintf(os.Stderr, "Error: --dsn is required\n")
		os.Exit(1)
	}

	// Determine source and destination
	var source string
	var destTable string

	if flag.NArg() == 1 {
		// Only destination table provided, must use --sql or --table for source
		if *sqlFlag == "" && *tableFlag == "" {
			fmt.Fprintf(os.Stderr, "Error: Either provide source file or use --sql/--table flag\n")
			os.Exit(1)
		}
		destTable = flag.Arg(0)
	} else {
		// Both source and destination provided
		source = flag.Arg(0)
		destTable = flag.Arg(1)
	}

	// Read data from source
	var records []map[string]any
	var err error

	if source != "" {
		// Read from file
		fmt.Fprintf(os.Stderr, "Reading from file: %s\n", source)
		records, err = fileiterator.ReadInput(source)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading source: %v\n", err)
			os.Exit(1)
		}
	} else if *sqlFlag != "" || *tableFlag != "" {
		// Read from SQL
		srcDSN := normalizeDSN(*driverFlag, *dsnFlag)
		sqlQuery := *sqlFlag
		if sqlQuery == "" && *tableFlag != "" {
			sqlQuery = fmt.Sprintf("SELECT * FROM %s", *tableFlag)
		}
		fmt.Fprintf(os.Stderr, "Executing source SQL: %s\n", sqlQuery)
		records, err = fileiterator.ReadSQLInput(*driverFlag, srcDSN, sqlQuery)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from SQL: %v\n", err)
			os.Exit(1)
		}
	}

	if len(records) == 0 {
		fmt.Fprintf(os.Stderr, "No records to insert\n")
		os.Exit(0)
	}

	fmt.Fprintf(os.Stderr, "Read %d records\n", len(records))

	// Connect to destination database
	dsn := normalizeDSN(*driverFlag, *dsnFlag)
	driver := *driverFlag
	if driver == "postgre" || driver == "postgresql" {
		driver = "postgres"
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "Error ping database: %v\n", err)
		os.Exit(1)
	}

	// Get all column names (sorted for consistency)
	columnSet := make(map[string]bool)
	for _, record := range records {
		for key := range record {
			columnSet[key] = true
		}
	}

	columns := make([]string, 0, len(columnSet))
	for col := range columnSet {
		columns = append(columns, col)
	}
	sort.Strings(columns)

	// Infer column types from data
	columnTypes := inferColumnTypes(records, columns)

	// Create table if not exists
	fmt.Fprintf(os.Stderr, "Creating table if not exists: %s\n", destTable)
	if err := createTableIfNotExists(db, destTable, columns, columnTypes, *driverFlag); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating table: %v\n", err)
		os.Exit(1)
	}

	// Insert data using BatchInserter
	fmt.Fprintf(os.Stderr, "Inserting %d records (batch size: %d)...\n", len(records), *batchFlag)

	fieldList := strings.Join(columns, ", ")
	insert, flush := hbsql.BatchInserter(db, destTable, fieldList, *batchFlag)
	defer flush()

	for _, record := range records {
		values := make([]any, len(columns))
		for i, col := range columns {
			values[i] = record[col]
		}
		insert(values)
	}

	flush() // Ensure all records are inserted

	fmt.Fprintf(os.Stderr, "Successfully inserted %d records into %s\n", len(records), destTable)
}

func inferColumnTypes(records []map[string]any, columns []string) map[string]string {
	types := make(map[string]string)

	for _, col := range columns {
		// Sample first non-nil value to infer type
		var sampleValue any
		for _, record := range records {
			if val, ok := record[col]; ok && val != nil {
				sampleValue = val
				break
			}
		}

		if sampleValue == nil {
			types[col] = "TEXT" // Default to TEXT for all-NULL columns
			continue
		}

		switch sampleValue.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			types[col] = "BIGINT"
		case float32, float64:
			types[col] = "DOUBLE"
		case bool:
			types[col] = "BOOLEAN"
		default:
			types[col] = "TEXT"
		}
	}

	return types
}

func createTableIfNotExists(db *sql.DB, tableName string, columns []string, columnTypes map[string]string, driver string) error {
	// Build CREATE TABLE statement
	var columnDefs []string
	for _, col := range columns {
		colType := columnTypes[col]

		// Adjust types for PostgreSQL
		if driver == "postgre" || driver == "postgres" || driver == "postgresql" {
			switch colType {
			case "DOUBLE":
				colType = "DOUBLE PRECISION"
			case "BOOLEAN":
				colType = "BOOLEAN"
			}
		}

		columnDefs = append(columnDefs, fmt.Sprintf("%s %s", col, colType))
	}

	createSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n  %s\n)",
		tableName,
		strings.Join(columnDefs, ",\n  "))

	_, err := db.Exec(createSQL)
	return err
}

func normalizeDSN(driver, dsn string) string {
	// If DSN contains "/" or "=" or "sslmode=", it's already in proper format
	if strings.Contains(dsn, "/") || strings.Contains(dsn, "=") || strings.Contains(dsn, "sslmode=") {
		return dsn
	}

	// Simplified format: "user:password@host"
	if !strings.Contains(dsn, "@") {
		return dsn
	}

	parts := strings.Split(dsn, "@")
	if len(parts) != 2 {
		return dsn
	}

	userPass := parts[0]
	host := parts[1]

	switch driver {
	case "mysql":
		if !strings.Contains(host, ":") {
			host = host + ":3306"
		}
		return fmt.Sprintf("%s@tcp(%s)/", userPass, host)

	case "postgre", "postgres", "postgresql":
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

func showHelp() {
	fmt.Fprintf(os.Stderr, "any2db - Import data from files or databases to database tables\n")
	fmt.Fprintf(os.Stderr, "================================================================\n\n")
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  File to DB:  %s --dsn=\"user:pass@host/dbname\" <source-file> <dest-table>\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  SQL to DB:   %s --dsn=\"user:pass@host/dbname\" --sql=\"SELECT...\" <dest-table>\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  Table copy:  %s --dsn=\"user:pass@host/dbname\" --table=\"source.table\" <dest-table>\n\n", os.Args[0])

	fmt.Fprintf(os.Stderr, "Flags:\n")
	fmt.Fprintf(os.Stderr, "  --dsn=\"connection-string\"    Destination database connection string (required)\n")
	fmt.Fprintf(os.Stderr, "  --sql=\"SELECT * FROM table\"  Source SQL query\n")
	fmt.Fprintf(os.Stderr, "  --table=\"schema.table\"       Source table name\n")
	fmt.Fprintf(os.Stderr, "  --driver=mysql               Database driver: mysql or postgre (default: mysql)\n")
	fmt.Fprintf(os.Stderr, "  --batch=1000                 Batch size for inserts (default: 1000)\n\n")

	fmt.Fprintf(os.Stderr, "Features:\n")
	fmt.Fprintf(os.Stderr, "  • Automatically creates destination table if not exists\n")
	fmt.Fprintf(os.Stderr, "  • Infers column types from data (BIGINT, DOUBLE, TEXT, BOOLEAN)\n")
	fmt.Fprintf(os.Stderr, "  • Supports MySQL and PostgreSQL\n")
	fmt.Fprintf(os.Stderr, "  • Uses batch inserts for performance\n")
	fmt.Fprintf(os.Stderr, "  • Auto-escapes values to prevent SQL injection\n\n")

	fmt.Fprintf(os.Stderr, "Supported source formats:\n")
	fmt.Fprintf(os.Stderr, "  • Parquet (.parquet, .pk)\n")
	fmt.Fprintf(os.Stderr, "  • JSONL (.jsonl, .ndjson)\n")
	fmt.Fprintf(os.Stderr, "  • CSV (.csv)\n")
	fmt.Fprintf(os.Stderr, "  • MsgPack (.msgpack, .mp)\n")
	fmt.Fprintf(os.Stderr, "  • SQL queries (via --sql or --table)\n")
	fmt.Fprintf(os.Stderr, "  • All formats support compression (.gz, .zst, .lz4, etc.)\n\n")

	fmt.Fprintf(os.Stderr, "Examples:\n")
	fmt.Fprintf(os.Stderr, "  # Import CSV file to MySQL table\n")
	fmt.Fprintf(os.Stderr, "  %s --dsn=\"root:pass@localhost/mydb\" data.csv users\n\n", os.Args[0])

	fmt.Fprintf(os.Stderr, "  # Import Parquet to PostgreSQL\n")
	fmt.Fprintf(os.Stderr, "  %s --driver=postgre --dsn=\"user:pass@pghost/mydb\" data.parquet public.orders\n\n", os.Args[0])

	fmt.Fprintf(os.Stderr, "  # Copy table from one DB to another (same server)\n")
	fmt.Fprintf(os.Stderr, "  %s --dsn=\"root:pass@localhost/destdb\" --table=\"sourcedb.users\" users_copy\n\n", os.Args[0])

	fmt.Fprintf(os.Stderr, "  # Import with SQL transformation\n")
	fmt.Fprintf(os.Stderr, "  %s --dsn=\"root:pass@localhost/mydb\" --sql=\"SELECT id, name, price*1.1 as new_price FROM products\" products_adjusted\n\n", os.Args[0])

	fmt.Fprintf(os.Stderr, "  # Import compressed JSONL with custom batch size\n")
	fmt.Fprintf(os.Stderr, "  %s --dsn=\"root:pass@localhost/mydb\" --batch=5000 data.jsonl.zst events\n\n", os.Args[0])

	fmt.Fprintf(os.Stderr, "Notes:\n")
	fmt.Fprintf(os.Stderr, "  • Destination table is created with inferred schema if not exists\n")
	fmt.Fprintf(os.Stderr, "  • If table exists, data is appended (columns must match)\n")
	fmt.Fprintf(os.Stderr, "  • Column names are sorted alphabetically\n")
	fmt.Fprintf(os.Stderr, "  • All string values are auto-escaped for security\n\n")

	fmt.Fprintf(os.Stderr, "See also: ./any2jsonl, ./any2parquet, ./any2csv\n")
}
