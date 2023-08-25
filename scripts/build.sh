#!/bin/bash

cd "$(dirname "$0")/.."
root_path=$(pwd)
cd "${root_path}/gost/main/cmd/gost/"

go build
mv gost /tmp/proxy_forwarder

cd "$root_path"
