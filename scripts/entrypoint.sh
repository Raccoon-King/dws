#!/bin/sh
# Iron Bank compliant entrypoint script
# All entrypoint scripts must be in scripts/ folder per Iron Bank requirements

# Set default values with backward compatibility
PORT=${PORT:-${DWS_PORT:-8080}}
RULES_FILE=${RULES_FILE:-${DWS_RULES_FILE:-/etc/dws/rules.yaml}}
DEBUG=${DEBUG:-${DWS_DEBUG:-false}}
LOGGING=${LOGGING:-stdout}
LLM_ENABLED=${LLM_ENABLED:-false}
LLM_CONFIG=${LLM_CONFIG:-/etc/dws/llm.yaml}

# Validate required files exist
if [ ! -f "/dws" ]; then
    echo "ERROR: DWS binary not found at /dws"
    exit 1
fi

# Health check mode
if [ "$1" = "-health-check" ]; then
    exec /dws -health-check
    exit $?
fi

# Log startup information to STDOUT (Iron Bank requirement)
echo "Starting Document Scanner Service (DWS)"
echo "Port: $PORT"
echo "Rules file: $RULES_FILE"
echo "Debug mode: $DEBUG"
echo "Logging: $LOGGING"
echo "LLM enabled: $LLM_ENABLED"
if [ "$LLM_ENABLED" = "true" ]; then
    echo "LLM config: $LLM_CONFIG"
fi

# Execute the main application with all arguments
exec /dws "$@"