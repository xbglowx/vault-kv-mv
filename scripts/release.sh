#!/bin/bash
set -euo pipefail

trap clean EXIT

function clean {
    test -z ${BUILD_DIR:+x} || rm -rf "$BUILD_DIR"
    rm -f *.tar.gz
}

if [ $# -ne 1 ]; then
    echo "Usage: $0 RELEASE"
    exit 1
fi

RELEASE=$1
BUILD_DIR=$(mktemp -d)
BINARY=vault-kv-mv
GOARCH=amd64
export GOARCH

: "${GITHUB_TOKEN:?Need to set environment variable GITHUB_TOKEN}"

OUTPUT=$(curl -s -XPOST \
    -H "Authorization: token $GITHUB_TOKEN" \
    -H "Content-Type: application/json" \
    --data "{\"tag_name\": \"v$RELEASE\"}" \
    https://api.github.com/repos/xbglowx/vault-kv-mv/releases
)
RELEASE_ID=$(echo "$OUTPUT" |jq -r '.id')

declare -a OSES=("linux" "darwin")
for os in "${OSES[@]}"; do
    TAR_FILENAME="vault-kv-mv-${RELEASE}.${os}-${GOARCH}.tar.gz"
    export GOOS=$os
    go build -o "$BUILD_DIR/$BINARY"
    tar -czvf "$TAR_FILENAME" -C "$BUILD_DIR" "$BINARY"
    curl -XPOST \
        -H "Authorization: token $GITHUB_TOKEN" \
        -H "Content-Type: $(file -b --mime-type "$TAR_FILENAME")" \
        --data-binary @"$TAR_FILENAME" \
        "https://uploads.github.com/repos/xbglowx/vault-kv-mv/releases/$RELEASE_ID/assets?name=$TAR_FILENAME"
done

