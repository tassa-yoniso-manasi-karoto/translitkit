# TranslitKit Commands & Guidelines

## Build & Test Commands
* Build project: `go build`
* Run code generation: `go generate`
* Run all tests: `go test ./...`
* Run single test: `go test github.com/tassa-yoniso-manasi-karoto/translitkit/lang/zho -run TestFunctionName`
* Run tests with verbose output: `go test -v ./...`
* Lint code: `go vet ./...` and `golint ./...`

## Code Style Guidelines
* **Imports**: Group standard library first, then third-party, alphabetically sorted within groups
* **Formatting**: Use gofmt/goimports, tabs for indentation
* **Types**: Interface-based design with provider pattern, descriptive type names
* **Naming**: CamelCase for exported items, camelCase for unexported, acronyms capitalized (e.g., ISO639)
* **Error Handling**: Use `fmt.Errorf("context: %w", err)` for wrapped errors with context
* **Comments**: GoDoc-style for exported functions, document parameters and return values
* **Testing**: Use `testing` with `testify/assert`, test edge cases
* **Organization**: Language-specific code in `lang/`, common utilities in `common/`