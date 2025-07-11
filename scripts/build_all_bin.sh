#!/bin/bash

set -euo pipefail

cd "$(dirname "$0")"

PATH_SCRIPT="$(pwd)"
PATH_REPO="${PATH_SCRIPT}/.."
PATH_BUILD="${PATH_REPO}/build"
FILE_BEG="proxy-forwarder-"

cd "$PATH_REPO"
mkdir -p "$PATH_BUILD"
rm -f "$PATH_BUILD"/*


function compile() {
    os="$1" arch="$2"
    echo ''
    echo ''
    echo "# COMPILING BINARIES FOR ${os}-${arch} #"
    echo ''

    GOOS="$os" GOARCH="$arch" bash "${PATH_SCRIPT}/build.sh"
    echo ''
    GOOS="$os" GOARCH="$arch" CGO_ENABLED=0 bash "${PATH_SCRIPT}/build.sh"

    echo ''
    echo '### COMPRESSING ###'
    file="${FILE_BEG}${os}-${arch}"
    cd "$PATH_BUILD"
    tar -czf "${file}.tar.gz" "$file"
    tar -czf "${file}-CGO0.tar.gz" "${file}-CGO0"
}

compile "linux" "amd64"

# untested
compile "linux" "386"
compile "linux" "arm"
compile "linux" "arm64"

compile "freebsd" "386"
compile "freebsd" "amd64"
compile "freebsd" "arm"

compile "openbsd" "386"
compile "openbsd" "amd64"
compile "openbsd" "arm"

compile "darwin" "amd64"
compile "darwin" "arm64"
