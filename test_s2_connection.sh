#!/bin/bash

# Get S2 server details from config
S2_HOST=$(grep -A 3 "s2_server" configs/config.json | grep "host" | cut -d'"' -f4)
S2_PORT=$(grep -A 3 "s2_server" configs/config.json | grep "port" | awk '{print $2}' | tr -d ',')

echo "Testing connection to S2 at $S2_HOST:$S2_PORT"

# Test basic connectivity
echo -n "1. Testing network connectivity: "
if ping -c 1 -W 2 $S2_HOST > /dev/null 2>&1; then
    echo "✓ OK"
else
    echo "✗ Failed - Cannot reach S2 server"
    exit 1
fi

# Test HTTP port
echo -n "2. Testing HTTP port $S2_PORT: "
if nc -z -w 2 $S2_HOST $S2_PORT > /dev/null 2>&1; then
    echo "✓ OK"
else
    echo "✗ Failed - Port $S2_PORT not accessible"
    exit 1
fi

# Test health endpoint
echo -n "3. Testing S2 health endpoint: "
HEALTH=$(curl -s --connect-timeout 5 http://$S2_HOST:$S2_PORT/health)
if [ "$HEALTH" == "OK" ]; then
    echo "✓ OK"
else
    echo "✗ Failed - Health check returned: $HEALTH"
    exit 1
fi

# Test a single call
echo -n "4. Testing single call submission: "
RESPONSE=$(curl -s -X POST http://$S2_HOST:$S2_PORT/process-incoming \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "uniqueid=test_$(date +%s)&callerid=19543004835&extension=50764137984")

if [[ "$RESPONSE" == *"SET VARIABLE"* ]]; then
    echo "✓ OK"
    echo "   Response from S2:"
    echo "$RESPONSE" | head -3
else
    echo "✗ Failed"
    echo "   Response: $RESPONSE"
fi

echo ""
echo "Connection test complete!"
