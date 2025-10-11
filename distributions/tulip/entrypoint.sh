#!/bin/sh

# Entrypoint script for Tulip OpenTelemetry Collector
# This script makes Tulip a drop-in replacement for upstream OTel Collector images
# by checking for config files in known locations used by core and contrib distributions.

set -e

# Default config file paths in order of preference:
# 1. Tulip's own config path (highest priority)
# 2. Contrib config path (otelcol-contrib compatibility)
# 3. Core config path (otelcol compatibility)
CONFIG_PATHS="/etc/tulip/config.yaml /etc/otelcol-contrib/config.yaml /etc/otelcol/config.yaml"

# Function to find the first existing config file
find_config() {
    for path in $CONFIG_PATHS; do
        if [ -f "$path" ]; then
            echo "$path"
            return 0
        fi
    done
    return 1
}

# If no --config flag is provided, find and use the first available config
if ! echo "$@" | grep -q -- "--config"; then
    if CONFIG_FILE=$(find_config); then
        echo "Using config file: $CONFIG_FILE"
        exec /tulip --config="$CONFIG_FILE" "$@"
    else
        echo "ERROR: No config file found in any of the following locations:"
        for path in $CONFIG_PATHS; do
            echo "  - $path"
        done
        exit 1
    fi
else
    # --config flag was provided, use it as-is
    exec /tulip "$@"
fi
