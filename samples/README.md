# Sample Data Files

Example data files demonstrating any2parquet's support for ANY record structure with automatic schema inference.

## Sample Datasets

### 1. Product Catalog (`products.jsonl`)

E-commerce product data with mixed types:

```jsonl
{"id":1,"product":"Laptop Pro 15","category":"Electronics","price":1299.99,"in_stock":true,"quantity":25,"rating":4.7}
```

**Schema (auto-inferred):**
- id (int64)
- product (string)
- category (string)
- price (float64)
- in_stock (bool)
- quantity (int64)
- rating (float64)

**Files:**
- `products.jsonl` (589 bytes) - Plain JSONL
- `products.jsonl.gz` (260 bytes) - JSONL + Gzip (56% smaller)
- `products.parquet` (1.6KB) - Parquet + Snappy
- `products.parquet.zst` (789 bytes) - Parquet + Zstandard (51% smaller)

### 2. IoT Sensor Data (`sensors.jsonl`)

Heterogeneous sensor readings with different field sets:

```jsonl
{"sensor_id":"temp-01","location":"server-room","temperature":22.5,"humidity":45,"online":true,"last_reading":1640995200}
{"sensor_id":"motion-01","location":"entrance","detections":142,"battery":87,"online":true,"last_reading":1640995200}
```

**Schema (auto-inferred, merged from all records):**
- sensor_id (string)
- location (string)
- temperature (float64, optional - not in motion sensors)
- humidity (int64, optional - not in motion sensors)
- detections (int64, optional - not in temp sensors)
- battery (int64, optional - not in temp sensors)
- online (bool)
- last_reading (int64)

**Files:**
- `sensors.jsonl` (594 bytes) - Plain JSONL
- `sensors.jsonl.gz` (241 bytes) - JSONL + Gzip (59% smaller)
- `sensors.parquet` (1.8KB) - Parquet + Snappy

### 3. User Data (`users.csv`)

CSV file with header row:

```csv
user_id,username,email,age,premium,credits,country
1001,alice_smith,alice@example.com,28,true,150.50,USA
```

**Schema (auto-inferred from CSV values):**
- user_id (int64)
- username (string)
- email (string)
- age (int64)
- premium (bool)
- credits (float64)
- country (string)

**Files:**
- `users.csv` (323 bytes) - Plain CSV
- `users.csv.zst` (220 bytes) - CSV + Zstandard (32% smaller)
- `users.parquet` (1.6KB) - Parquet + Snappy

## Compression Comparison

| Format | Size | Compression | Notes |
|--------|------|-------------|-------|
| JSONL (plain) | 589B | None | Human-readable |
| JSONL + Gzip | 260B | 56% smaller | Good compression, slow |
| Parquet + Snappy | 1.6KB | Built-in | Fast read/write, columnar |
| Parquet + Zstd | 789B | 51% smaller | Best compression for Parquet |

**Recommendation:** Use plain `.parquet` (Snappy compression built-in). Only add `.zst` if you need absolute smallest size.

## Usage Examples

### Convert JSONL to Parquet
```bash
any2parquet products.jsonl products.parquet
```

### Convert compressed input
```bash
any2parquet products.jsonl.gz products.parquet
```

### Convert CSV to Parquet
```bash
any2parquet users.csv users.parquet
```

### Add extra compression to Parquet
```bash
any2parquet products.jsonl products.parquet.zst
```

### Test with different schemas
```bash
# Product catalog schema
any2parquet products.jsonl test1.parquet

# Sensor data schema (different fields!)
any2parquet sensors.jsonl test2.parquet

# CSV with inferred types
any2parquet users.csv test3.parquet
```

## Schema Inference

The converter automatically:
1. **Detects field names** from your data (alphabetically sorted for consistency)
2. **Infers types** from values:
   - Numbers without decimal → int64
   - Numbers with decimal → float64
   - "true"/"false" → bool
   - Everything else → string
3. **Handles missing fields** (sets to null/default)
4. **Merges schemas** across records (union of all fields)

## File Formats

All samples demonstrate the converter's ability to handle:
- **Different schemas** (products vs sensors vs users)
- **Different input formats** (JSONL, CSV)
- **Different compressions** (gzip, zstd)
- **Type inference** (automatic int64/float64/bool/string detection)

The output is always Parquet with Snappy compression (or additional compression if requested).
