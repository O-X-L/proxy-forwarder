#!/bin/bash

set -eo pipefail

cd "$(dirname "$0")"
PATH_REPO="$(pwd)/.."

GO_VERSION="$(head -n3 < "${PATH_REPO}/go.mod" | tail -n1 | cut -d ' ' -f2)"

if [ -z "$GO_BIN" ]
then
  GO_BIN='go'
fi

if [ -z "$GOOS" ]
then
  GOOS='linux'
fi

if [ -z "$GOARCH" ]
then
  GOARCH='amd64'
fi

cgo=''
if [ -n "$CGO_ENABLED" ]
then
  cgo="-CGO${CGO_ENABLED}"
fi

if ! $GO_BIN version | grep -q "$GO_VERSION"
then
  echo "ERROR: GO is not of required version '${GO_VERSION}'!"
  exit 1
fi

set -u

cd "$(dirname "$0")/.."
PATH_BUILD="$(pwd)/build"
mkdir -p "$PATH_BUILD"

echo '### DOWNLOADING DEPENDENCIES ###'
$GO_BIN mod tidy

FILE_BUILD="${PATH_BUILD}/proxy-forwarder-${GOOS}-${GOARCH}${cgo}"

echo ''
echo '### BUILDING ###'
$GO_BIN build -o "$FILE_BUILD" ./gost/main/cmd/gost/

echo ''
echo "DONE: ${FILE_BUILD}"
