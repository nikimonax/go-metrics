#!/bin/bash

set -euo pipefail

BIN_DIR=bin

if [ -f .env ]; then
    set -o allexport
    source .env
    set +o allexport
else
    echo "[WARN] .env file not exists"
fi

if [ -z "${1:-}" ]; then
    START=1
    ITER="$(git branch --show-current | sed -n 's/^iter\([0-9]\+\)$/\1/p')"
else
    START="$1"
    ITER="$1"
fi

if [ -z "$ITER" ]; then
    echo "[ERROR]: Test iteration could not be determined."
    echo "Please provide 'ITER' variable."
    exit 1
fi

if (( ITER >= 10 )); then
    # начиная с 10 инкремента нужна тестовая база данных
    if ! nc -z -w 3 "$DATABASE_HOST" "$DATABASE_PORT" &>/dev/null; then
        echo "[ERROR]: Database not running"
        echo "Run in docker using 'make up'"
        exit 1
    fi
fi

DATABASE_DSN="postgres://$DATABASE_USER:$DATABASE_PASSWORD@$DATABASE_HOST:$DATABASE_PORT/$DATABASE_NAME"

for ((i=START; i<=ITER; i++)); do
    echo -n "Iteration $i: "

    # начиная с 7 инкремента используем api с json
    if (( i < 7 )); then
        export API=1
    else
        export API=2
    fi

    "$BIN_DIR/metricstest" \
        -test.run="^TestIteration$i[AB]*$" \
        -binary-path="$BIN_DIR/server" \
        -agent-binary-path="$BIN_DIR/agent" \
        -server-port="$(( 8000 + RANDOM % 1000 ))" \
        -source-path="." \
        -file-storage-path=$(mktemp) \
        -database-dsn="$DATABASE_DSN"
done