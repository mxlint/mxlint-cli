#!/bin/sh

OPA="./bin/opa"

if [ ! -f "$OPA" ]; then
    echo "Program not found, downloading..."
    curl -L -o "$OPA" https://openpolicyagent.org/downloads/v0.62.1/opa_darwin_amd64
    chmod +x "$OPA"
fi

