package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/parf/homebase-go-lib/fileiterator"
)

func main() {
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "any2parquet - Convert any format to Parquet (RECOMMENDED) 🏆\n")
		fmt.Fprintf(os.Stderr, "=============================================================\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s <input-file> [output-file]\n\n", os.Args[0])

		fmt.Fprintf(os.Stderr, "Supported input formats (recognized extension → format):\n")
		fmt.Fprintf(os.Stderr, "  - JSONL: JSON Lines, one JSON object per line → .jsonl\n")
		fmt.Fprintf(os.Stderr, "  - CSV: Comma-separated values with header row → .csv, .tsv, .psv\n")
		fmt.Fprintf(os.Stderr, "  - MsgPack: Binary serialization format → .msgpack\n")
		fmt.Fprintf(os.Stderr, "  - FlatBuffer: Zero-copy binary format → .fb\n\n")

		fmt.Fprintf(os.Stderr, "Input compression formats (recognized extension → format):\n")
		fmt.Fprintf(os.Stderr, "  .gz  → Gzip (standard compression, widely supported, slow)\n")
		fmt.Fprintf(os.Stderr, "  .zst → Zstandard (RECOMMENDED: best balance of speed & compression)\n")
		fmt.Fprintf(os.Stderr, "  .lz4 → LZ4 (fastest compression, moderate compression ratio)\n")
		fmt.Fprintf(os.Stderr, "  .br  → Brotli (best compression, but very slow)\n")
		fmt.Fprintf(os.Stderr, "  .xz  → XZ/LZMA (excellent compression, extremely slow - avoid)\n\n")

		fmt.Fprintf(os.Stderr, "What is Parquet (.parquet)?\n")
		fmt.Fprintf(os.Stderr, "  Parquet is a columnar storage format optimized for analytics.\n")
		fmt.Fprintf(os.Stderr, "  Winner in benchmarks: Fastest overall (0.61s), excellent compression (44MB).\n")
		fmt.Fprintf(os.Stderr, "  Best for: Everything - APIs, analytics, data warehouses.\n")
		fmt.Fprintf(os.Stderr, "  Compatible with: Spark, DuckDB, Pandas, Arrow, all major data tools.\n\n")

		fmt.Fprintf(os.Stderr, "⚠️  IMPORTANT: Parquet already has built-in Snappy compression!\n")
		fmt.Fprintf(os.Stderr, "   Additional compression (.parquet.gz/.zst/.lz4) is usually NOT needed.\n")
		fmt.Fprintf(os.Stderr, "   It only gives ~10-15%% smaller files but slower access.\n\n")

		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  %s data.jsonl.gz                      → data.parquet\n", os.Args[0])
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

		fmt.Fprintf(os.Stderr, "Note: Assumes record structure with fields:\n")
		fmt.Fprintf(os.Stderr, "  id, name, email, age, score, active, category, timestamp\n\n")

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
		for _, ext := range []string{".jsonl", ".csv", ".msgpack", ".fb"} {
			outputFile = strings.TrimSuffix(outputFile, ext)
		}
		outputFile += ".parquet"
	}

	fmt.Printf("Converting %s -> %s\n", inputFile, outputFile)

	// Read all records from input
	records, err := readInput(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Read %d records\n", len(records))

	// Write to Parquet (compression auto-detected from filename)
	if err := fileiterator.WriteParquet(outputFile, records); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing Parquet: %v\n", err)
		os.Exit(1)
	}

	stat, _ := os.Stat(outputFile)
	fmt.Printf("Written %s (%d bytes, %.2f MB)\n", outputFile, stat.Size(), float64(stat.Size())/1024/1024)
}

func readInput(filename string) ([]fileiterator.ParquetRecord, error) {
	lower := strings.ToLower(filename)

	if strings.Contains(lower, ".fb") {
		return readFlatBuffer(filename)
	} else if strings.Contains(lower, ".msgpack") {
		return readMsgPack(filename)
	} else if strings.Contains(lower, ".csv") {
		return readCSV(filename)
	} else if strings.Contains(lower, ".jsonl") {
		return readJSONL(filename)
	}

	return nil, fmt.Errorf("unsupported input format: %s", filename)
}

func readJSONL(filename string) ([]fileiterator.ParquetRecord, error) {
	var records []fileiterator.ParquetRecord
	err := fileiterator.IterateJSONL(filename, func(line map[string]any) error {
		rec := fileiterator.ParquetRecord{
			ID:        int64(line["id"].(float64)),
			Name:      line["name"].(string),
			Email:     line["email"].(string),
			Age:       int64(line["age"].(float64)),
			Score:     line["score"].(float64),
			Active:    line["active"].(bool),
			Category:  line["category"].(string),
			Timestamp: int64(line["timestamp"].(float64)),
		}
		records = append(records, rec)
		return nil
	})
	return records, err
}

func readCSV(filename string) ([]fileiterator.ParquetRecord, error) {
	var records []fileiterator.ParquetRecord

	reader := fileiterator.FUOpen(filename)
	defer reader.Close()

	csvReader := csv.NewReader(reader)

	// Detect delimiter
	if strings.Contains(filename, ".tsv") {
		csvReader.Comma = '\t'
	} else if strings.Contains(filename, ".psv") {
		csvReader.Comma = '|'
	}

	// Skip header
	_, err := csvReader.Read()
	if err != nil {
		return nil, err
	}

	for {
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		var rec fileiterator.ParquetRecord
		fmt.Sscanf(row[0], "%d", &rec.ID)
		rec.Name = row[1]
		rec.Email = row[2]
		fmt.Sscanf(row[3], "%d", &rec.Age)
		fmt.Sscanf(row[4], "%f", &rec.Score)
		rec.Active = row[5] == "true"
		rec.Category = row[6]
		fmt.Sscanf(row[7], "%d", &rec.Timestamp)

		records = append(records, rec)
	}

	return records, nil
}

func readMsgPack(filename string) ([]fileiterator.ParquetRecord, error) {
	var records []fileiterator.ParquetRecord
	err := fileiterator.IterateMsgPack(filename, func(data any) error {
		m := data.(map[string]any)
		rec := fileiterator.ParquetRecord{
			ID:        m["id"].(int64),
			Name:      m["name"].(string),
			Email:     m["email"].(string),
			Age:       m["age"].(int64),
			Score:     m["score"].(float64),
			Active:    m["active"].(bool),
			Category:  m["category"].(string),
			Timestamp: m["timestamp"].(int64),
		}
		records = append(records, rec)
		return nil
	})
	return records, err
}

func readFlatBuffer(filename string) ([]fileiterator.ParquetRecord, error) {
	var records []fileiterator.ParquetRecord
	err := fileiterator.IterateFlatBufferList(filename, func(data []byte) error {
		// Note: Simplified - would need full FlatBuffer parsing with generated code
		return fmt.Errorf("FlatBuffer reading not yet fully implemented - use jsonl2parquet or csv2parquet instead")
	})
	return records, err
}
