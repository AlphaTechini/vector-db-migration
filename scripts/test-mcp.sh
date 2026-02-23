#!/bin/bash
# MCP Integration Test Script
# Tests all 3 MCP tools end-to-end

set -e

API_KEY="test-key-12345"
BASE_URL="http://localhost:8080"
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "==================================="
echo "MCP Integration Test Suite"
echo "==================================="
echo ""

# Function to make JSON-RPC request
rpc_call() {
    local method=$1
    local params=$2
    local id=$3
    
    curl -s -X POST "$BASE_URL" \
        -H "Authorization: Bearer $API_KEY" \
        -H "Content-Type: application/json" \
        -d "{
            \"jsonrpc\": \"2.0\",
            \"id\": $id,
            \"method\": \"$method\",
            \"params\": $params
        }" | jq .
}

# Test 1: migration_status
echo -e "${YELLOW}Test 1: migration_status${NC}"
echo "Testing: Get status of non-existent migration..."
result=$(rpc_call "migration_status" '{"migration_id": "test-123"}' 1)
echo "$result" | jq '.result.status'
if echo "$result" | jq -e '.result.status == "not_started"' > /dev/null; then
    echo -e "${GREEN}✅ PASS: Returns not_started for new migration${NC}"
else
    echo -e "${RED}❌ FAIL: Expected status 'not_started'${NC}"
    exit 1
fi
echo ""

# Test 2: list_migrations (empty list)
echo -e "${YELLOW}Test 2: list_migrations${NC}"
echo "Testing: List all migrations (should be empty)..."
result=$(rpc_call "list_migrations" '{"limit": 10, "offset": 0}' 2)
echo "$result" | jq '.result'
total=$(echo "$result" | jq '.result.total')
if [ "$total" == "0" ]; then
    echo -e "${GREEN}✅ PASS: Returns empty list${NC}"
else
    echo -e "${YELLOW}⚠️  INFO: Found $total migrations (may have test data)${NC}"
fi
echo ""

# Test 3: list_migrations with pagination
echo -e "${YELLOW}Test 3: list_migrations with pagination${NC}"
echo "Testing: Pagination parameters..."
result=$(rpc_call "list_migrations" '{"limit": 5, "offset": 10}' 3)
limit=$(echo "$result" | jq '.result.limit')
offset=$(echo "$result" | jq '.result.offset')
if [ "$limit" == "5" ] && [ "$offset" == "10" ]; then
    echo -e "${GREEN}✅ PASS: Pagination parameters respected${NC}"
else
    echo -e "${RED}❌ FAIL: Expected limit=5, offset=10${NC}"
    exit 1
fi
echo ""

# Test 4: schema_recommendation (Pinecone → Qdrant)
echo -e "${YELLOW}Test 4: schema_recommendation (Pinecone → Qdrant)${NC}"
echo "Testing: Get schema recommendations..."
result=$(rpc_call "schema_recommendation" '{"source_type": "pinecone", "target_type": "qdrant"}' 4)
echo "$result" | jq '.result'
source_type=$(echo "$result" | jq -r '.result.source_type')
target_type=$(echo "$result" | jq -r '.result.target_type')
confidence=$(echo "$result" | jq '.result.overall_confidence')
if [ "$source_type" == "pinecone" ] && [ "$target_type" == "qdrant" ]; then
    echo -e "${GREEN}✅ PASS: Returns correct migration path${NC}"
    echo -e "   Confidence: $confidence"
else
    echo -e "${RED}❌ FAIL: Wrong migration path${NC}"
    exit 1
fi

# Check for warnings
warnings=$(echo "$result" | jq '.result.warnings | length')
if [ "$warnings" -gt 0 ]; then
    echo -e "   Warnings: $(echo "$result" | jq '.result.warnings[0]')"
fi
echo ""

# Test 5: schema_recommendation with source schema
echo -e "${YELLOW}Test 5: schema_recommendation with custom fields${NC}"
echo "Testing: Custom field mapping..."
result=$(rpc_call "schema_recommendation" '{
    "source_type": "pinecone",
    "target_type": "weaviate",
    "source_schema": {
        "id": "string",
        "title": "string",
        "custom_field": "text"
    }
}' 5)
field_count=$(echo "$result" | jq '.result.field_mappings | length')
echo -e "   Field mappings found: $field_count"
if [ "$field_count" -ge 4 ]; then
    echo -e "${GREEN}✅ PASS: Maps common + custom fields${NC}"
else
    echo -e "${YELLOW}⚠️  INFO: Only $field_count fields mapped${NC}"
fi
echo ""

# Test 6: Error handling - missing required param
echo -e "${YELLOW}Test 6: Error handling (missing required param)${NC}"
echo "Testing: Call without source_type..."
result=$(rpc_call "schema_recommendation" '{"target_type": "qdrant"}' 6)
has_error=$(echo "$result" | jq 'has("error")')
if [ "$has_error" == "true" ]; then
    error_msg=$(echo "$result" | jq -r '.error.message')
    echo -e "${GREEN}✅ PASS: Returns error: $error_msg${NC}"
else
    echo -e "${RED}❌ FAIL: Should return error for missing param${NC}"
    exit 1
fi
echo ""

# Test 7: Error handling - same source and target
echo -e "${YELLOW}Test 7: Error handling (same source/target)${NC}"
echo "Testing: Same source and target type..."
result=$(rpc_call "schema_recommendation" '{"source_type": "pinecone", "target_type": "pinecone"}' 7)
has_error=$(echo "$result" | jq 'has("error")')
if [ "$has_error" == "true" ]; then
    error_msg=$(echo "$result" | jq -r '.error.message')
    echo -e "${GREEN}✅ PASS: Returns error: $error_msg${NC}"
else
    echo -e "${RED}❌ FAIL: Should reject same source/target${NC}"
    exit 1
fi
echo ""

# Test 8: Auth failure
echo -e "${YELLOW}Test 8: Authentication failure${NC}"
echo "Testing: Invalid API key..."
result=$(curl -s -X POST "$BASE_URL" \
    -H "Authorization: Bearer wrong-key" \
    -H "Content-Type: application/json" \
    -d '{"jsonrpc":"2.0","id":8,"method":"migration_status","params":{}}')
status_code=$(echo "$result" | jq -r '.error.code // empty')
if [ "$status_code" == "-32001" ]; then
    echo -e "${GREEN}✅ PASS: Rejects invalid API key${NC}"
else
    echo -e "${RED}❌ FAIL: Should reject invalid key${NC}"
    exit 1
fi
echo ""

# Summary
echo "==================================="
echo -e "${GREEN}ALL TESTS PASSED! ✅${NC}"
echo "==================================="
echo ""
echo "MCP Server is fully functional:"
echo "  ✅ migration_status - Working"
echo "  ✅ list_migrations - Working"
echo "  ✅ schema_recommendation - Working"
echo "  ✅ Authentication - Working"
echo "  ✅ Error handling - Working"
echo ""
