# AGENTS.MD - gossh Development Guidelines

## Build Commands

### Building the Project
```bash
# Build the executable
go build -o gossh .

# Build with specific version
./build.sh v1.0.0

# Cross-platform builds (using build script)
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o gossh-linux-amd64 .
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o gossh-darwin-amd64 .
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o gossh-windows-amd64.exe
```

### Running the Application
```bash
# After building
./gossh

# Or run directly without building
go run main.go
```

## Lint Commands

### Standard Go Tools
```bash
# Run gofmt to format code
gofmt -w .

# Run golint
golint ./...

# Run go vet for static analysis
go vet ./...

# Run all checks
go fmt ./...
go vet ./...
```

### Using golangci-lint (if installed)
```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run

# Run linter with specific configuration
golangci-lint run --config .golangci.yml
```

## Test Commands

### Running Tests
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Run a specific test file
go test ./path/to/package -run ^TestFunctionName$

# Run tests in a specific package
go test ./config/...
go test ./ssh/...
```

### Test Structure
Since there are no test files currently, when adding tests, follow Go conventions:
- Name test files with `_test.go` suffix
- Use `func TestXxx(t *testing.T)` function signatures
- Place tests in the same package as the code being tested

## Code Style Guidelines

### Imports
- Group imports with blank lines between groups
- Standard library imports first
- Third-party libraries next
- Local project imports last

```go
import (
    "fmt"
    "os"

    "github.com/pkg/sftp"
    "golang.org/x/crypto/ssh"

    "gossh/config"
)
```

### Formatting
- Use tabs for indentation (not spaces)
- Maximum line length of 120 characters
- Run `gofmt -w .` to format all files
- One space after control statements (if, for, switch)

### Naming Conventions
- Use camelCase for exported functions and variables
- Use PascalCase for exported types
- Use snake_case for internal/CLI flag variables
- Be descriptive but concise in naming
- Acronyms should be capitalized (SSH, JSON, URL, not Url)

### Types and Structs
- Define structs with meaningful field names
- Use struct tags for JSON serialization
- Document public structs with comments

```go
type Connection struct {
    Name            string `json:"name"`                // Unique identifier for the connection
    Group           string `json:"group,omitempty"`     // Optional group assignment
    User            string `json:"user"`                // Username for SSH connection
    Host            string `json:"host"`                // Host address
    Port            int    `json:"port"`                // SSH port (default 22)
    KeyPath         string `json:"key_path,omitempty"`  // Path to SSH private key
    CredentialAlias string `json:"credential_alias,omitempty"` // Alias for stored password
}
```

### Error Handling
- Always check and handle errors appropriately
- Use fmt.Errorf with descriptive messages
- Include underlying errors when wrapping them
- Use specific error types when appropriate

```go
if err != nil {
    return fmt.Errorf("failed to create session: %w", err)
}
```

### Comments and Documentation
- Document exported functions, types, and variables
- Use complete sentences in comments
- Include examples when complex logic is involved
- Use // for single-line comments
- Use /* */ for multi-line comments when appropriate

### Security Considerations
- Store sensitive information encrypted (as implemented in credentials.go)
- Validate user inputs before using them
- Sanitize file paths to prevent directory traversal
- Use secure defaults for SSH configurations
- Never log sensitive information (passwords, private keys)

### Best Practices
- Follow the DRY principle (Don't Repeat Yourself)
- Keep functions focused on a single responsibility
- Use context for cancellation and timeouts when applicable
- Prefer early returns to reduce nesting
- Handle edge cases gracefully
- Use defer for cleanup operations
- Close resources (files, connections) properly

### Dependency Management
- Use Go modules (already configured in go.mod)
- Keep dependencies updated
- Only add necessary dependencies
- Review security vulnerabilities in dependencies regularly

### Git Workflow
- Use feature branches for new functionality
- Write clear, descriptive commit messages
- Follow conventional commits format when possible
- Run tests before committing
- Keep commits atomic (one logical change per commit)