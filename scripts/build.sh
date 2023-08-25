#!/bin/bash

OUT_BIN='/tmp/proxy_forwarder'

cd "$(dirname "$0")/.."
root_path=$(pwd)
cd "${root_path}/gost/main/cmd/gost/"

go build
mv gost "$OUT_BIN"

cd "$root_path"

echo "Binary created: '${OUT_BIN}'"
