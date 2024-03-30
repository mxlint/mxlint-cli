#!/bin/sh

set -e

OPA="./bin/opa"

if [ ! -f "$OPA" ]; then
    echo "Program not found, downloading..."
    curl -L -o "$OPA" https://openpolicyagent.org/downloads/v0.63.0/opa_darwin_amd64
    chmod +x "$OPA"
fi

OPA test -v policies

