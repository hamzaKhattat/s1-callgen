#!/bin/bash

echo "Starting S1 Call Generator Test..."

# Check if S2 is reachable
if ! nc -z 10.0.0.2 5060 2>/dev/null; then
    echo "Error: S2 (10.0.0.2:5060) is not reachable"
    exit 1
fi

# Start with test configuration
cd /home/car/s1-callgen

# Create a test config with lower call rate
cat > configs/test_config.json << 'JSON'
{
    "s2_server": {
        "host": "10.0.0.2",
        "port": 5060
    },
    "call_params": {
        "acd_min": 10,
        "acd_max": 30,
        "asr": 80.0,
        "max_concurrent": 5,
        "calls_per_second": 0.5,
        "ramp_up_time": 10,
        "ramp_down_time": 10
    },
    "schedule": {
        "weekday": {
            "start_hour": 0,
            "end_hour": 24
        },
        "weekend": {
            "start_hour": 0,
            "end_hour": 24
        }
    }
}
JSON

# Run the generator
./bin/callgen -config configs/test_config.json -csv configs/numbers.csv
