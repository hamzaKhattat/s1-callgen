#!/bin/bash

echo "Testing S2 Call Router..."

# Test health endpoint
echo -n "1. Testing health endpoint: "
HEALTH=$(curl -s http://localhost:8000/health)
if [ "$HEALTH" == "OK" ]; then
    echo "✓ OK"
else
    echo "✗ Failed"
fi

# Test stats endpoint
echo -n "2. Testing stats endpoint: "
STATS=$(curl -s http://localhost:8000/stats)
if [ $? -eq 0 ]; then
    echo "✓ OK"
    echo "   Stats: $(echo $STATS | jq -c .)"
else
    echo "✗ Failed"
fi

# Test incoming call with proper form data
echo -n "3. Testing incoming call processing: "
RESPONSE=$(curl -s -X POST http://localhost:8000/process-incoming \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "uniqueid=test_call_123&callerid=19543004835&extension=50764137984")

if [[ "$RESPONSE" == *"SET VARIABLE"* ]]; then
    echo "✓ OK"
    echo "   Response:"
    echo "$RESPONSE" | head -5
else
    echo "✗ Failed"
    echo "   Response: $RESPONSE"
fi

# Test with multiple calls
echo ""
echo "4. Testing multiple calls:"
for i in {1..5}; do
    CALL_ID="test_call_$i_$(date +%s)"
    ANI="1954300483$i"
    DNIS="5076413798$i"
    
    echo -n "   Call $i (ID: $CALL_ID): "
    
    RESPONSE=$(curl -s -X POST http://localhost:8000/process-incoming \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "uniqueid=$CALL_ID&callerid=$ANI&extension=$DNIS")
    
    if [[ "$RESPONSE" == *"SET VARIABLE"* ]]; then
        echo "✓ Success"
        # Extract DID from response
        DID=$(echo "$RESPONSE" | grep "SET VARIABLE DID_ASSIGNED" | awk '{print $4}')
        echo "      DID assigned: $DID"
    else
        echo "✗ Failed"
    fi
    
    sleep 1
done

# Check updated stats
echo ""
echo "5. Checking updated statistics:"
STATS=$(curl -s http://localhost:8000/stats)
echo "   Active calls: $(echo $STATS | jq -r .active_calls)"
echo "   Total DIDs: $(echo $STATS | jq -r .total_dids)"
echo "   In use DIDs: $(echo $STATS | jq -r .in_use_dids)"

echo ""
echo "Test complete!"
