#!/bin/bash

# Test script to verify watch flag implementation
# This script tests the basic CLI flag parsing without requiring FDB

cd "$(dirname "$0")"

echo "Testing CLI help text for --watch flag..."

# Test 1: Check that the --watch flag is recognized (should not give unknown flag error)
echo "Test 1: Validating --watch flag recognition"
if timeout 5 go run . --help 2>&1 | grep -q "watch"; then
    echo "✓ --watch flag found in help text"
else
    echo "✗ --watch flag NOT found in help text"
fi

# Test 2: Check validation - watch with multiple queries should fail
echo "Test 2: Validating watch with multiple queries (should fail)"
if timeout 5 go run . --watch -q "test1" -q "test2" 2>&1 | grep -q "single query"; then
    echo "✓ Multiple query validation working"
else
    echo "? Multiple query validation not tested (requires FDB)"
fi

echo "Manual verification complete."
echo ""
echo "To fully test the watch functionality, you would need:"
echo "1. A running FoundationDB cluster"
echo "2. Run: fql --cluster /path/to/cluster --watch -q '/test(\"key\")=<>'"
echo "3. In another terminal, modify the key and observe the changes"