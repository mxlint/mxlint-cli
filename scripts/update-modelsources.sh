#!/bin/sh
#
#
echo "
modelsource: resources/modelsource-v1
projectDirectory: resources/app-mpr-v1
" > /tmp/mxlint-dev.yaml
./bin/mxlint-darwin-arm64 export-model --config /tmp/mxlint-dev.yaml

echo "
modelsource: resources/modelsource-v2
projectDirectory: resources/app-mpr-v2
" > /tmp/mxlint-dev.yaml
./bin/mxlint-darwin-arm64 export-model --config /tmp/mxlint-dev.yaml
