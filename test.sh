#!/bin/bash
set -e

echo "========================================="
echo "Testing any2parquet and parquet2db"
echo "========================================="
echo

# Cleanup function
cleanup() {
    rm -f test_*.parquet test_*.jsonl test_output.* 2>/dev/null || true
}
trap cleanup EXIT

# Test 1: Convert JSONL to Parquet
echo "Test 1: Convert JSONL to Parquet"
./any2parquet samples/products.jsonl test_products.parquet
if [ -f test_products.parquet ]; then
    SIZE=$(stat -f%z test_products.parquet 2>/dev/null || stat -c%s test_products.parquet)
    echo "✓ Created test_products.parquet ($SIZE bytes)"
else
    echo "✗ Failed to create test_products.parquet"
    exit 1
fi
echo

# Test 2: Convert compressed CSV to Parquet
echo "Test 2: Convert compressed CSV to Parquet"
./any2parquet samples/users.csv.zst test_users.parquet
if [ -f test_users.parquet ]; then
    SIZE=$(stat -f%z test_users.parquet 2>/dev/null || stat -c%s test_users.parquet)
    echo "✓ Created test_users.parquet ($SIZE bytes)"
else
    echo "✗ Failed to create test_users.parquet"
    exit 1
fi
echo

# Test 3: Convert Parquet to stdout (pipe test)
echo "Test 3: Convert to stdout and verify output"
LINES=$(./any2parquet samples/sensors.jsonl - | wc -c)
if [ "$LINES" -gt 0 ]; then
    echo "✓ Stdout output works ($LINES bytes)"
else
    echo "✗ Stdout output failed"
    exit 1
fi
echo

# Test 4: parquet2db - only if database is available
echo "Test 4: parquet2db database import"
if [ -n "$TEST_DSN" ]; then
    ./parquet2db --dsn="$TEST_DSN" test_products.parquet test_products_table
    echo "✓ Imported to database (DSN: $TEST_DSN)"
else
    echo "⊘ Skipped (set TEST_DSN to test database imports)"
    echo "  Example: export TEST_DSN=\"root:password@localhost/testdb\""
fi
echo

# Test 5: SQL export - only if database is available
echo "Test 5: any2parquet SQL export"
if [ -n "$TEST_DSN" ] && [ -n "$TEST_SQL" ]; then
    ./any2parquet --dsn="$TEST_DSN" --sql="$TEST_SQL" test_sql_output.parquet
    if [ -f test_sql_output.parquet ]; then
        SIZE=$(stat -f%z test_sql_output.parquet 2>/dev/null || stat -c%s test_sql_output.parquet)
        echo "✓ SQL export successful ($SIZE bytes)"
    else
        echo "✗ SQL export failed"
        exit 1
    fi
else
    echo "⊘ Skipped (set TEST_DSN and TEST_SQL to test SQL exports)"
    echo "  Example: export TEST_SQL=\"SELECT * FROM users LIMIT 10\""
fi
echo

echo "========================================="
echo "All tests passed! ✓"
echo "========================================="
