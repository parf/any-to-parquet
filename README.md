# any-to-parquet

Universal converter to Parquet format. Convert from JSONL, CSV, MsgPack, or FlatBuffer to Parquet with automatic compression detection.

## 🏆 Why Parquet?

Parquet is the winner in our benchmarks: fastest overall (0.61s), excellent compression (44MB for 1M records), and industry-standard format compatible with Spark, DuckDB, Pandas, Arrow, and all major data tools.

## Installation

```bash
go install github.com/parf/any-to-parquet/cmd/any2parquet@latest
```

## Usage

```bash
# Basic conversion
any2parquet data.jsonl.gz                      # → data.parquet
any2parquet data.csv.zst output.pq             # → output.pq

# With additional compression (usually not needed)
any2parquet data.jsonl data.parquet.zst        # → data.parquet.zst (with Zstd)
any2parquet data.msgpack data.parquet.lz4      # → data.parquet.lz4 (with LZ4)
```

**⚠️ Important:** Parquet already has built-in Snappy compression. Additional compression (.parquet.gz/.zst/.lz4) gives minimal benefit (~10-15% smaller) but slower access.

## Supported Input Formats

- **JSONL** (.jsonl) - JSON Lines, one object per line
- **CSV** (.csv, .tsv, .psv) - Comma/tab/pipe separated values  
- **MsgPack** (.msgpack) - Binary serialization
- **FlatBuffer** (.fb) - Zero-copy binary format

Input compression: .gz, .zst, .lz4, .br, .xz (auto-detected)

## Performance (1M records)

- **Size:** 44MB (72% smaller than plain text)
- **Read:** 0.15s (4x faster than MsgPack, 13x faster than JSONL)
- **Write:** 0.46s (fastest binary format)

Full benchmark results: https://github.com/parf/homebase-go-lib/blob/main/benchmarks/serialization-benchmark-result.md

## Building from Source

```bash
git clone https://github.com/parf/any-to-parquet.git
cd any-to-parquet
go build ./cmd/any2parquet
```

## License

MIT
