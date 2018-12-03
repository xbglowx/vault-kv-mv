# vault-kv-mv [![CircleCI](https://circleci.com/gh/xbglowx/vault-kv-mv.svg?style=svg)](https://circleci.com/gh/xbglowx/vault-kv-mv)
Easily move Hashicorp Vault keys to different paths

## Build
1. clone this repo and step into the dir
1. `go get -d .`
1. `go build vault-kv-mv.go`

## Test
1. go test

## Usage
`./vault-kv-mv <source_key_path> <destination_key_path>`
