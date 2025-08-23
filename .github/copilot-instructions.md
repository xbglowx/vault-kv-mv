# Copilot Instructions for vault-kv-mv

## Project Overview

vault-kv-mv is a Go command-line tool that simplifies moving HashiCorp Vault secrets between different paths. It's designed to help administrators reorganize their Vault secret stores efficiently and safely.

**Key Features:**
- Move individual secrets between paths
- Move entire directories of secrets 
- Support for both trailing slash and non-trailing slash path conventions
- Safe operations: creates new entries before deleting old ones
- Built for Vault KV v1 secrets engine

## Development Guidelines

### Go Code Standards

This project follows standard Go conventions:

- Use `gofmt` for code formatting
- Follow effective Go naming conventions (camelCase for unexported, PascalCase for exported)
- Keep functions focused and single-purpose
- Use descriptive variable names, especially for Vault paths and operations
- Prefer explicit error handling over panic()

### Architecture Patterns

- **vaultClient struct**: Wraps Vault API client with domain-specific methods
- **Path manipulation**: Core logic handles 4 scenarios for source/destination combinations (file→file, file→dir, dir→dir, dir→file)
- **Functional approach**: Pure functions for path transformations make testing easier
- **Fail-fast**: Validate inputs and Vault connectivity before making changes

### Vault-Specific Considerations

**Security & Safety:**
- Always read secrets before attempting to write to validate existence
- Write new secrets before deleting old ones to prevent data loss
- Use Vault's logical client for KV operations
- Be mindful that this tool works with KV v1 engine (no versioning)

**Path Handling:**
- Trailing slashes matter: they distinguish files from directories
- The `OldNewPaths()` function handles 4 path combination scenarios
- Use `path.Base()` for extracting filenames from paths

### Testing Patterns

This project uses testcontainers for integration testing:

```go
// Standard test setup pattern
client, closer := testVaultServer(t)
defer closer()
```

**Test Structure:**
- Each test uses a fresh Vault container
- Tests validate both successful operations and cleanup (old paths deleted)
- Use meaningful test data that reflects real-world scenarios
- Test all 4 path combination scenarios

**When adding tests:**
- Use descriptive test names that explain the scenario
- Always test both positive outcomes and error conditions
- Verify old paths are properly cleaned up after moves
- Use the existing `testVaultServer()` helper for consistency

### Build & Development Workflow

**Building:**
```bash
go get -d .
go build vault-kv-mv.go
```

**Testing:**
```bash
go test -v
```

**Code Quality:**
- The project uses GitHub Actions for CI/CD
- golangci-lint enforces code quality standards
- CodeQL performs security analysis

### Common Development Tasks

**Adding new path scenarios:**
1. Update `OldNewPaths()` function logic
2. Add corresponding test cases
3. Ensure error handling for edge cases

**Modifying Vault operations:**
1. Always test with real Vault containers
2. Consider error scenarios (network failures, permission issues)
3. Maintain the fail-safe pattern (write before delete)

**Error handling:**
- Use `log.Fatalf()` for unrecoverable errors that should stop execution
- Provide clear, actionable error messages
- Include relevant context (paths, operation being performed)

## File Structure

- `vault-kv-mv.go` - Main application logic
- `vault-kv-mv_test.go` - Integration tests using testcontainers
- `.github/workflows/` - CI/CD pipelines
- `scripts/release.sh` - Release automation

## Dependencies

Key dependencies to be aware of:
- `github.com/hashicorp/vault/api` - Official Vault Go client
- `github.com/testcontainers/testcontainers-go` - Testing with real Vault instances

## Security Considerations

- This tool requires Vault authentication (token-based)
- Operations are destructive (deletes old paths after successful moves)
- Always test path manipulations thoroughly before production use
- Consider implementing dry-run functionality for safety

## Maintenance Notes

⚠️ This project is currently looking for a maintainer. When contributing:
- Keep changes minimal and well-tested
- Maintain backward compatibility
- Follow the existing patterns and conventions
- Add comprehensive tests for any new functionality