# Contributing to ModbusGo

Thank you for your interest in contributing to ModbusGo! We welcome contributions from the community and are grateful for your support.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [How to Contribute](#how-to-contribute)
- [Development Setup](#development-setup)
- [Coding Standards](#coding-standards)
- [Testing Guidelines](#testing-guidelines)
- [Commit Messages](#commit-messages)
- [Pull Request Process](#pull-request-process)
- [Reporting Issues](#reporting-issues)

## Code of Conduct

By participating in this project, you agree to abide by our Code of Conduct:

- Be respectful and inclusive
- Welcome newcomers and help them get started
- Focus on constructive criticism
- Accept feedback gracefully
- Prioritize the project's best interests

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/modbusgo.git
   cd modbusgo
   ```
3. **Add upstream remote**:
   ```bash
   git remote add upstream https://github.com/adibhanna/modbus-go.git
   ```
4. **Create a branch** for your changes:
   ```bash
   git checkout -b feature/your-feature-name
   ```

## How to Contribute

### Types of Contributions

- **Bug Fixes**: Fix issues reported in GitHub Issues
- **Features**: Add new functionality or enhance existing features
- **Documentation**: Improve or add documentation
- **Tests**: Add missing tests or improve test coverage
- **Performance**: Optimize code for better performance
- **Refactoring**: Improve code quality without changing functionality

### Before You Start

1. Check if an issue already exists for your contribution
2. For major changes, open an issue first to discuss
3. Ensure your idea aligns with the project's goals

## Development Setup

### Prerequisites

- Go 1.20 or higher
- Git
- Make (optional but recommended)

### Setting Up Your Environment

1. **Install Go** from [golang.org](https://golang.org/dl/)

2. **Install development tools**:
   ```bash
   make install-tools
   ```
   
   Or manually:
   ```bash
   go install golang.org/x/tools/cmd/goimports@latest
   go install github.com/securego/gosec/v2/cmd/gosec@latest
   curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
   ```

3. **Install dependencies**:
   ```bash
   make deps
   ```
   
   Or:
   ```bash
   go mod download
   ```

4. **Run tests to verify setup**:
   ```bash
   make test
   ```

## Coding Standards

### Go Code Style

- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` and `goimports` for formatting:
  ```bash
  make fmt
  ```
- Follow naming conventions:
  - Exported names start with capital letters
  - Acronyms should be all caps (HTTP, TCP, RTU)
  - Use descriptive variable names

### Project-Specific Guidelines

1. **Error Handling**:
   ```go
   // Good
   if err != nil {
       return nil, fmt.Errorf("failed to read registers: %w", err)
   }
   
   // Bad
   if err != nil {
       return nil, err
   }
   ```

2. **Type Safety**:
   ```go
   // Good
   var address modbus.Address = 100
   
   // Bad
   var address uint16 = 100
   ```

3. **Constants**:
   - Define protocol constants in `modbus/constants.go`
   - Group related constants together
   - Add comments for non-obvious values

4. **Documentation**:
   - All exported types and functions must have comments
   - Comments should start with the name being declared
   - Include examples for complex functions

### Linting

Run linters before committing:

```bash
make lint
```

Fix common issues:
- Unused variables and imports
- Error handling
- Code complexity
- Security issues

## Testing Guidelines

### Test Requirements

- All new features must include tests
- Bug fixes should include regression tests
- Maintain or improve test coverage

### Types of Tests

1. **Unit Tests**:
   ```go
   func TestReadHoldingRegisters(t *testing.T) {
       // Test implementation
   }
   ```

2. **Integration Tests**:
   ```go
   // +build integration
   
   func TestClientServerIntegration(t *testing.T) {
       // Integration test
   }
   ```

3. **Benchmarks**:
   ```go
   func BenchmarkReadRegisters(b *testing.B) {
       // Benchmark implementation
   }
   ```

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make coverage

# Run benchmarks
make bench

# Run specific tests
go test -run TestReadCoils ./...
```

### Test Best Practices

- Use table-driven tests for multiple scenarios
- Test error cases and edge conditions
- Use mock interfaces for external dependencies
- Keep tests focused and independent
- Use meaningful test names

## Commit Messages

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- **feat**: New feature
- **fix**: Bug fix
- **docs**: Documentation changes
- **style**: Code style changes (formatting, etc.)
- **refactor**: Code refactoring
- **test**: Test additions or changes
- **perf**: Performance improvements
- **chore**: Maintenance tasks

### Examples

```bash
feat(client): add support for RTU over TCP

Implements RTU over TCP transport as specified in the
MODBUS specification. This allows RTU frame format to
be used over TCP/IP networks.

Closes #123
```

```bash
fix(server): correct CRC calculation for RTU frames

The CRC was being calculated with bytes in wrong order.
This fix ensures proper byte ordering as per specification.
```

## Pull Request Process

### Before Submitting

1. **Update your branch**:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Run checks**:
   ```bash
   make pre-commit
   ```
   
   This runs:
   - Formatting checks
   - Linting
   - Tests
   - Vet

3. **Update documentation** if needed

4. **Add tests** for your changes

### Submitting a Pull Request

1. **Push your changes**:
   ```bash
   git push origin feature/your-feature-name
   ```

2. **Create Pull Request** on GitHub

3. **Fill out the PR template** completely:
   - Description of changes
   - Related issues
   - Testing performed
   - Breaking changes (if any)

4. **Address review feedback** promptly

### Pull Request Guidelines

- Keep PRs focused and small
- One feature or fix per PR
- Include tests
- Update documentation
- Ensure CI passes
- Respond to reviews within 48 hours

## Reporting Issues

### Before Creating an Issue

1. Search existing issues
2. Check the documentation
3. Verify with latest version

### Creating an Issue

Include the following information:

1. **Version** of ModbusGo
2. **Go version** and OS
3. **Description** of the problem
4. **Steps to reproduce**
5. **Expected behavior**
6. **Actual behavior**
7. **Code sample** (if applicable)
8. **Error messages** or logs

### Issue Template

```markdown
## Description
Brief description of the issue

## Steps to Reproduce
1. Step one
2. Step two
3. Step three

## Expected Behavior
What should happen

## Actual Behavior
What actually happens

## Environment
- ModbusGo version: v1.0.0
- Go version: 1.21
- OS: Ubuntu 22.04

## Code Sample
```go
// Your code here
```

## Error Output
```
Error messages or logs
```
```

## Development Workflow

### Typical Workflow

1. **Pick an issue** or create one
2. **Fork and clone** the repository
3. **Create a branch** from `main`
4. **Make changes** with tests
5. **Run checks**: `make pre-commit`
6. **Commit** with meaningful messages
7. **Push** to your fork
8. **Create PR** and wait for review
9. **Address feedback** if needed
10. **Merge** after approval

### Using Make Commands

```bash
# Development cycle
make fmt          # Format code
make test         # Run tests
make lint         # Check code quality
make coverage     # Check test coverage

# Before committing
make pre-commit   # Run all checks

# Build and run
make build        # Build library
make examples     # Build examples
make run-tcp-server  # Run server example

# Utilities
make clean        # Clean build artifacts
make deps         # Update dependencies
make help         # Show all commands
```

## Getting Help

- **Documentation**: Read [DOCUMENTATION.md](DOCUMENTATION.md)
- **API Reference**: Check [API_REFERENCE.md](API_REFERENCE.md)
- **GitHub Issues**: Ask questions with the "question" label
- **Discussions**: Use GitHub Discussions for general topics

## Recognition

Contributors will be:
- Listed in the project's contributors section
- Mentioned in release notes for significant contributions
- Given credit in commit messages and PRs

## License

By contributing, you agree that your contributions will be licensed under the same MIT License that covers this project.

Thank you for contributing to ModbusGo! 🎉