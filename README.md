# ⚠️ Looking for a maintainer ⚠️
Looking for someone to take this project from me. https://github.com/xbglowx/vault-kv-mv/issues/94


# vault-kv-mv [![Build and Test](https://github.com/xbglowx/vault-kv-mv/actions/workflows/build-test.yaml/badge.svg)](https://github.com/xbglowx/vault-kv-mv/actions/workflows/build-test.yaml) [![CodeQL](https://github.com/xbglowx/vault-kv-mv/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/xbglowx/vault-kv-mv/actions/workflows/codeql-analysis.yml) [![golangci-lint](https://github.com/xbglowx/vault-kv-mv/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/xbglowx/vault-kv-mv/actions/workflows/golangci-lint.yml)
Easily move Hashicorp Vault keys to different paths

## Build
1. clone this repo and step into the dir
1. `go get -d .`
1. `go build vault-kv-mv.go`

## Test
1. go test

## Usage
`./vault-kv-mv <source_key_path> <destination_key_path>`
