# 🚀 any-to-parquet

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![GitHub Stars](https://img.shields.io/github/stars/parf/any-to-parquet?style=social)](https://github.com/parf/any-to-parquet/stargazers)

**Lightning-fast universal data converter** | Transform JSONL, CSV, MsgPack, SQL databases → Apache Parquet with automatic compression detection

Convert any structured data format to Apache Parquet in seconds. Export from MySQL/PostgreSQL databases. Import to databases. Perfect for data pipelines, ETL workflows, analytics, and big data processing with Spark, DuckDB, Pandas, and Arrow.

---

## 🎯 Why Parquet?

Parquet is the **industry-standard columnar format** that delivers:

| Feature | Benefit |
|---------|---------|
| 🏆 **Performance** | 13x faster reads than JSONL |
| 💾 **Compression** | 72% smaller files (44MB vs 156MB) |
| 🔧 **Compatibility** | Works with Spark, DuckDB, Pandas, Arrow, Snowflake |
| 📊 **Analytics** | Optimized for column-based queries |
| ⚡ **Speed** | 0.15s read, 0.46s write for 1M records |
| 📝 **Extensions** | Supports both `.parquet` and `.pk` (shorter alternative) |

Full benchmark: [Performance Comparison](https://github.com/parf/homebase-go-lib/blob/main/benchmarks/serialization-benchmark-result.md)

---

## 📦 Installation

### Quick Install (Recommended)

```bash
go install github.com/parf/any-to-parquet@latest
```

This installs both `any2parquet` and `parquet2db` utilities.

### Build from Source

```bash
git clone https://github.com/parf/any-to-parquet.git
cd any-to-parquet
go build -o any2parquet main.go
go build -o parquet2db parquet2db.go
```

### Test Installation

```bash
# Run automated tests
./test.sh

# Quick smoke test
any2parquet samples/products.jsonl test.parquet
parquet2db --dsn="your-db-connection" test.parquet test_table
```

---

## 🚀 Quick Start

### File Conversion

```bash
# Convert JSONL to Parquet
any2parquet data.jsonl                  # → data.parquet

# Convert CSV to Parquet
any2parquet users.csv users.parquet

# Convert compressed files (auto-detect)
any2parquet logs.jsonl.gz              # → logs.parquet
any2parquet metrics.csv.zst            # → metrics.parquet

# Output to stdout
any2parquet data.csv - | duckdb
```

### SQL Database Export

```bash
# Export from MySQL
any2parquet --dsn="user:pass@localhost" --sql="SELECT * FROM users" users.parquet

# Export from PostgreSQL
any2parquet --driver=postgre --dsn="user:pass@host" --table="public.orders" orders.parquet

# Export to stdout and pipe
any2parquet --dsn="root:pass@localhost" --sql="SELECT * FROM logs LIMIT 100" - | head
```

### Database Import

```bash
# Import Parquet to MySQL
parquet2db --dsn="root:pass@localhost/mydb" users.parquet users_table

# Import to PostgreSQL
parquet2db --driver=postgre --dsn="user:pass@host/mydb" orders.parquet public.orders

# Automatic schema creation and batch inserts
parquet2db --dsn="root:pass@localhost/mydb" --batch=5000 large_file.parquet events
```

### Advanced Usage

```bash
# Add extra compression (optional)
any2parquet data.jsonl data.parquet.zst        # Parquet + Zstandard
any2parquet data.msgpack data.parquet.lz4      # Parquet + LZ4

# Custom output name
any2parquet input.csv output_name.parquet
```

> ⚠️ **Note:** Parquet has built-in Snappy compression. Additional compression (.zst/.lz4/.gz) provides only ~10-15% size reduction with slower access times.

---

## 📋 Supported Formats

### Input Formats

| Format | Extensions | Description |
|--------|-----------|-------------|
| 📄 **JSONL** | `.jsonl`, `.ndjson` | JSON Lines - one JSON object per line |
| 📊 **CSV** | `.csv`, `.tsv`, `.psv` | Comma/Tab/Pipe separated values |
| 🔧 **MsgPack** | `.msgpack`, `.mp` | Binary serialization format |
| 🗄️ **SQL Databases** | MySQL, PostgreSQL | Direct export via `--dsn` flag |

### Compression Support (Auto-Detected)

All input formats support automatic compression detection:

| Compression | Extension | Speed | Ratio | Use Case |
|-------------|-----------|-------|-------|----------|
| ⚡ **LZ4** | `.lz4` | Fastest | Good | Real-time processing |
| 🎯 **Zstandard** | `.zst` | Fast | Excellent | **Recommended** for most uses |
| 📦 **Gzip** | `.gz` | Slow | Good | Wide compatibility |
| 🔥 **Brotli** | `.br` | Very slow | Best | Maximum compression |
| ❄️ **XZ** | `.xz` | Extremely slow | Excellent | Avoid for most uses |

---

## ✨ Key Features

### any2parquet - Universal Converter
- ✅ **Universal Schema Support** - Works with ANY data structure
- ✅ **SQL Database Export** - Direct export from MySQL and PostgreSQL
- ✅ **Automatic Type Inference** - Detects int64, float64, string, bool
- ✅ **Zero Configuration** - No schema files needed
- ✅ **Compression Detection** - Automatically handles .gz, .zst, .lz4, etc.
- ✅ **Stdout Support** - Pipe data with `-` argument
- ✅ **Fast Processing** - Optimized for large datasets
- ✅ **Industry Standard** - Compatible with all major data tools

### parquet2db - Database Importer
- ✅ **Auto Schema Creation** - Creates tables automatically if not exists
- ✅ **Type Inference** - Intelligent column type detection (BIGINT, DOUBLE, TEXT, BOOLEAN)
- ✅ **Batch Inserts** - High-performance bulk imports (configurable batch size)
- ✅ **SQL Injection Protection** - Automatic value escaping
- ✅ **Multi-Database** - MySQL and PostgreSQL support

---

## 🗄️ SQL Database Support

### Export from Databases (any2parquet)

Export data directly from MySQL or PostgreSQL to Parquet format:

```bash
# MySQL export
any2parquet --dsn="user:pass@localhost" --sql="SELECT * FROM users" users.parquet
any2parquet --dsn="root:pass@localhost" --table="mydb.orders" orders.parquet

# PostgreSQL export
any2parquet --driver=postgre --dsn="user:pass@host" --sql="SELECT * FROM logs" logs.parquet
any2parquet --driver=postgre --dsn="user:pass@host" --table="public.events" events.parquet

# To stdout for piping
any2parquet --dsn="user:pass@host" --sql="SELECT * FROM users" - | duckdb
```

**DSN Format:**
- MySQL: `user:password@tcp(host:3306)/database` or simplified `user:password@host`
- PostgreSQL: `host=localhost port=5432 user=x password=y dbname=z sslmode=disable` or simplified `user:password@host`

### Import to Databases (parquet2db)

Import Parquet files to database tables with automatic schema creation:

```bash
# Import to MySQL
parquet2db --dsn="root:pass@localhost/mydb" data.parquet users_table

# Import to PostgreSQL
parquet2db --driver=postgre --dsn="user:pass@host/mydb" data.parquet public.events

# Custom batch size for large imports
parquet2db --dsn="root:pass@localhost/mydb" --batch=5000 large_file.parquet events

# Copy data between databases
parquet2db --dsn="root:pass@localhost/mydb" --table="source.users" target_users
```

**Features:**
- Automatically creates table if not exists
- Infers column types from data (BIGINT, DOUBLE, TEXT, BOOLEAN)
- Batch inserts for high performance (default: 1000 records)
- Auto-escapes values to prevent SQL injection

---

## 💡 Use Cases

### 🔄 Data Pipeline
Convert logs and events to Parquet for efficient storage and querying:
```bash
# Convert application logs
any2parquet app-logs.jsonl.gz logs.parquet

# Process with DuckDB
duckdb -c "SELECT * FROM 'logs.parquet' WHERE status = 500"
```

### 📊 Analytics Workflow
Transform CSV exports for data analysis:
```bash
# Convert CSV export
any2parquet sales-data.csv sales.parquet

# Analyze with pandas
python -c "import pandas as pd; df = pd.read_parquet('sales.parquet'); print(df.describe())"
```

### 🚀 ETL Processing
Optimize data lake storage:
```bash
# Convert streaming data
any2parquet events.jsonl.zst warehouse/events.parquet

# Query with Spark
spark-sql -e "SELECT * FROM parquet.\`warehouse/events.parquet\` LIMIT 10"
```

### 🗄️ Database Migration
Export and import between databases:
```bash
# Export from MySQL to Parquet
any2parquet --dsn="user:pass@oldserver" --table="mydb.users" users.parquet

# Import to PostgreSQL
parquet2db --driver=postgre --dsn="user:pass@newserver/newdb" users.parquet public.users

# Or pipe directly (future feature)
# any2parquet --dsn="mysql://old" --table="users" - | parquet2db --dsn="postgres://new" - users
```

---

## 📈 Performance

Based on 1 million records benchmark:

| Metric | Plain JSONL | JSONL.zst | **Parquet** | Improvement |
|--------|-------------|-----------|---------|-------------|
| File Size | 156 MB | 43 MB | **44 MB** | 72% smaller |
| Read Time | 1.93s | 1.91s | **0.15s** | **13x faster** |
| Write Time | 1.38s | 0.84s | **0.46s** | **3x faster** |

**Winner:** Parquet delivers the best balance of speed, compression, and compatibility.

---

## 🔧 Schema Support

### Automatic Type Detection

```jsonl
{"id": 1, "name": "Alice", "score": 98.5, "active": true}
{"id": 2, "name": "Bob", "score": 87.3, "active": false}
```

Automatically inferred as:
- `id` → int64
- `name` → string
- `score` → float64
- `active` → bool

### Any Structure Works

```csv
product_id,product_name,price,in_stock
101,Widget,15.99,true
102,Gadget,25.50,false
```

```bash
any2parquet products.csv products.parquet
# ✅ Converts with automatic schema detection
```

---

## 🛠️ Technical Details

### Built With
- **Go 1.21+** - High-performance language
- **Apache Arrow** - Columnar data format
- **homebase-go-lib** - Optimized I/O and compression

### Keywords
parquet converter, data format converter, jsonl to parquet, csv to parquet, msgpack to parquet, mysql to parquet, postgresql to parquet, database export, database import, sql to parquet, parquet to sql, parquet tool, apache parquet, columnar format, data pipeline, etl tool, big data, data engineering, analytics, database migration, spark compatible, duckdb compatible, pandas compatible

### Related Projects
- [homebase-go-lib](https://github.com/parf/homebase-go-lib) - Core library with format converters
- [Apache Parquet](https://parquet.apache.org/) - Official Parquet documentation
- [Apache Arrow](https://arrow.apache.org/) - In-memory columnar format

---

## 🤝 Contributing

Contributions welcome! Feel free to:
- 🐛 Report bugs
- 💡 Suggest features
- 🔧 Submit pull requests

---

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details

---

## ⭐ Star History

If you find this tool useful, please consider [giving it a star](https://github.com/parf/any-to-parquet/stargazers)! ⭐

[![Star History Chart](https://api.star-history.com/svg?repos=parf/any-to-parquet&type=Date)](https://star-history.com/#parf/any-to-parquet&Date)

---

<div align="center">

**Made with ❤️ for the data engineering community**

[Report Bug](https://github.com/parf/any-to-parquet/issues) · [Request Feature](https://github.com/parf/any-to-parquet/issues)

</div>
