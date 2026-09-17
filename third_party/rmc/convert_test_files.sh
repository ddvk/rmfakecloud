#!/bin/bash

set -euo pipefail

# Directory containing .rm files
TEST_DIR="tests/rm"
OUTPUT_DIR="test_output"

# Check if directory exists
if [ ! -d "$TEST_DIR" ]; then
    echo "Error: Directory $TEST_DIR does not exist"
    exit 1
fi

if [ ! -d "$OUTPUT_DIR" ]; then
    mkdir -p "$OUTPUT_DIR"
    mkdir -p "$OUTPUT_DIR/markdown"
    mkdir -p "$OUTPUT_DIR/svg"
    mkdir -p "$OUTPUT_DIR/pdf"
fi

# Iterate through all .rm files in the directory
for file in "$TEST_DIR"/*.rm; do
    # Check if files exist (in case directory is empty)
    if [ -f "$file" ]; then
        echo "Testing file: $file"
        file_name=$(basename "$file")
        
        # Run first test command
        echo "Running test markdown..."
        rmc -t markdown "$file" -o "$OUTPUT_DIR/markdown/$file_name.md"
        
        # Run second test command
        echo "Running test svg..."
        rmc -t svg "$file" -o "$OUTPUT_DIR/svg/$file_name.svg"

        # Run third test command
        echo "Running test pdf..."
        rmc -t pdf "$file" -o "$OUTPUT_DIR/pdf/$file_name.pdf"

        echo "----------------------------------------"
    fi
done

echo "All tests completed"
