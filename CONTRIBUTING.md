# Contributing to Task Manager

Thank you for considering contributing to the Task Manager project!

## Development Setup

1. Fork the repository
2. Clone your fork
3. Install dependencies: `make install`
4. Create a feature branch: `git checkout -b feature/my-feature`
5. Make your changes
6. Run tests: `make test`
7. Commit your changes: `git commit -am 'Add some feature'`
8. Push to the branch: `git push origin feature/my-feature`
9. Create a Pull Request

## Code Style

- Follow Go best practices and idioms
- Use `gofmt` for formatting
- Run `make fmt` before committing
- Write tests for new features
- Maintain test coverage above 70%

## Testing

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run integration tests
make integration-test
```

## Commit Messages

Use clear and descriptive commit messages:
- `feat: add new feature`
- `fix: resolve bug in handler`
- `docs: update README`
- `test: add unit tests for service`
- `refactor: improve code structure`

## Pull Request Process

1. Update README.md with details of changes if needed
2. Update tests to reflect your changes
3. Ensure all tests pass
4. Request review from maintainers

## Code of Conduct

Please be respectful and constructive in all interactions.

## License

By contributing, you agree that your contributions will be licensed under the same license as the project.
