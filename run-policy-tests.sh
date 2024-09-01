#!/bin/sh

set -e

OPA="./bin/opa"

UNAME="$(uname -s)"
echo "OS: $UNAME"

if [ "$UNAME" = "Linux" ]; then
    OPA_DL="opa_linux_amd64"
elif [ "$UNAME" = "Darwin" ]; then
    OPA_DL="opa_darwin_amd64"
else
    echo "Unsupported OS"
    exit 1
fi

if [ ! -f "$OPA" ]; then
    echo "Program not found, downloading..."
    mkdir -p bin
    curl -s -L -o "$OPA" "https://openpolicyagent.org/downloads/v0.63.0/$OPA_DL"
    chmod +x "$OPA"
fi

$OPA test -v policies
