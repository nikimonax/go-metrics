#!/bin/bash

set -euo pipefail

BIN_DIR=bin

if [ -z "${1:-}" ]; then
    ITER="$(git branch --show-current | sed -n 's/^iter\([0-9]\+\)$/\1/p')"
else
    ITER="$1"
fi

if [ -z "$ITER" ]; then
    echo "Test iteration could not be determined."
    echo "Please provide 'ITER' variable."
    exit 1
fi

for ((i=1; i<=ITER; i++)); do
    echo -n "Iteration $i: "

    # начиная с 7 инкремента используем api с json
    if (( ITER < 7 )); then
        export API=1
    else
        export API=2
    fi

    "$BIN_DIR/metricstest" \
        -test.run="^TestIteration$i[AB]*$" \
        -binary-path="$BIN_DIR/server" \
        -agent-binary-path="$BIN_DIR/agent" \
        -server-port="$(( 8000 + RANDOM % 1000 ))" \
        -source-path="."
done