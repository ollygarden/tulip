#!/bin/bash

set -e

DOCKER_IMAGE="${DOCKER_IMAGE:-tulip:local}"
CONTAINER_NAME="tulip-docker-test-$$"

while getopts d: flag
do
    case "${flag}" in
        d) distribution=${OPTARG};;
    esac
done

if [[ -z $distribution ]]; then
    echo "Distribution to test not provided. Use '-d' to specify the distribution. Ex.:"
    echo "$0 -d tulip"
    exit 1
fi

TEST_CONFIG="./distributions/${distribution}/${distribution}-test-docker.yaml"
if [[ ! -f "$TEST_CONFIG" ]]; then
    echo "❌ FAIL. Docker test config not found: ${TEST_CONFIG}"
    exit 1
fi

# cleanup on exit
function teardown {
    echo "🔧 Tearing down..."
    docker logs "${CONTAINER_NAME}" > ./test/logs/otelcol-docker-${distribution}.log 2>&1 || true
    docker rm -f "${CONTAINER_NAME}" > /dev/null 2>&1 || true
    rm -f ./test/test-docker-output.json

    echo "🪵 '${distribution}' Docker container logs"
    cat ./test/logs/otelcol-docker-${distribution}.log 2>/dev/null || true
}
trap teardown EXIT

mkdir -p ./test/logs

## start the container
echo "🔧 Starting '${distribution}' Docker container (image: ${DOCKER_IMAGE})..."
docker run -d \
    --name "${CONTAINER_NAME}" \
    -p 4317:4317 \
    -p 13133:13133 \
    -v "$(pwd)/${TEST_CONFIG}:/etc/tulip/config.yaml:ro" \
    -v "$(pwd)/test:/output" \
    "${DOCKER_IMAGE}" > /dev/null

## wait for health check
max_retries=50
retries=0
while true
do
    if ! docker inspect "${CONTAINER_NAME}" --format='{{.State.Running}}' 2>/dev/null | grep -q true; then
        echo "❌ FAIL. Container is not running."
        exit 1
    fi

    if curl -s localhost:13133 2>/dev/null | grep -q "Server available"; then
        echo "✅ '${distribution}' container is healthy."
        break
    fi

    retries=$((retries + 1))
    if [ "$retries" -gt "$max_retries" ]; then
        echo "❌ FAIL. Health check not ready after ~5s."
        exit 1
    fi
    sleep 0.1s
done

## generate traces
echo "🔧 Generating traces..."
go run github.com/open-telemetry/opentelemetry-collector-contrib/cmd/telemetrygen@latest \
    traces --otlp-endpoint localhost:4317 --otlp-insecure --traces 5 > /dev/null 2>&1

if [ $? != 0 ]; then
    echo "❌ FAIL. Failed to generate traces."
    exit 1
fi
echo "✅ Traces sent."

## wait briefly for the collector to flush
sleep 2

## check that traces were written
OUTPUT_FILE="./test/test-docker-output.json"
if [[ -f "$OUTPUT_FILE" ]] && [[ -s "$OUTPUT_FILE" ]]; then
    echo "✅ Traces received and written to output file."
else
    echo "❌ FAIL. No trace output found at ${OUTPUT_FILE}."
    exit 1
fi

echo "✅ PASS: '${distribution}' Docker image test"
exit 0
