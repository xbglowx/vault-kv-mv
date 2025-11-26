# ⚠️ Looking for a maintainer ⚠️
Looking for someone to take this project from me. https://github.com/xbglowx/vault-kv-mv/issues/94

# vault-kv-mv

[![Build and Test](https://github.com/xbglowx/vault-kv-mv/actions/workflows/build-test.yaml/badge.svg)](https://github.com/xbglowx/vault-kv-mv/actions/workflows/build-test.yaml) [![CodeQL](https://github.com/xbglowx/vault-kv-mv/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/xbglowx/vault-kv-mv/actions/workflows/codeql-analysis.yml) [![golangci-lint](https://github.com/xbglowx/vault-kv-mv/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/xbglowx/vault-kv-mv/actions/workflows/golangci-lint.yml)

`vault-kv-mv` is a command-line tool that simplifies moving and renaming secrets within HashiCorp Vault's Key-Value (KV) secrets engine. It supports moving single secrets, as well as recursively moving all secrets under a given path.

## Features

- **Move a single secret:** Rename a secret to a new path.
- **Move a secret into a directory:** Move a secret to be under a new path (directory).
- **Recursively move a directory:** Move all secrets from one path to another.

## Installation

### From source

If you have Go installed, you can build and install `vault-kv-mv` with the following commands:

```bash
git clone https://github.com/xbglowx/vault-kv-mv.git
cd vault-kv-mv
go install
```

### Pre-compiled binaries

Pre-compiled binaries for various operating systems are available on the [GitHub Releases page](https://github.com/xbglowx/vault-kv-mv/releases).

## Usage

The tool uses the standard HashiCorp Vault environment variables for authentication:
- `VAULT_ADDR`: The address of your Vault server.
- `VAULT_TOKEN`: Your Vault authentication token.

Make sure these are set before running the tool.

```bash
export VAULT_ADDR="http://127.0.0.1:8200"
export VAULT_TOKEN="s.xxxxxxxxxxxx"
```

The basic syntax is:

```bash
vault-kv-mv <source_path> <destination_path>
```

### Examples

#### 1. Rename a secret

To rename a secret from `secret/foo` to `secret/bar`:

```bash
vault-kv-mv secret/foo secret/bar
```

#### 2. Move a secret into a new directory

To move the secret `secret/foo` to `secret/new/foo`:

```bash
vault-kv-mv secret/foo secret/new/
```

#### 3. Move all secrets from one directory to another

To move all secrets from `secret/old/` to `secret/new/`:

```bash
vault-kv-mv secret/old/ secret/new/
```

## Development

### Building

1. Clone this repository.
2. `go build`

### Testing

Run the test suite with:

```bash
go test -v ./...
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
