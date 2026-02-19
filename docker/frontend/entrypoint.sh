#!/bin/sh
set -e

# Start api-gateway in background
/app/api-gateway &
GATEWAY_PID=$!

# Wait for api-gateway to be ready
for i in $(seq 1 30); do
    if wget -q -O /dev/null http://127.0.0.1:8080/health 2>/dev/null; then
        break
    fi
    if [ $i -eq 30 ]; then
        echo "api-gateway failed to start"
        exit 1
    fi
    sleep 0.5
done

# Start nginx in foreground (replaces shell)
exec nginx -g "daemon off;"
