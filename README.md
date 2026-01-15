# go-utils

A lightweight Go utilities library providing helpful string manipulation functions.

## Features

### String Utilities (`strutil`)

- **`IsEmpty(s *string) bool`** - Check if a string pointer is nil or empty

  - Returns `true` if the string is `nil` or has zero length

- **`ToSnakeCase(str string) string`** - Convert strings to snake_case format
  - Converts camelCase and PascalCase to snake_case
  - Example: `HelloWorld` → `hello_world`

## Usage

```go
import "github.com/yourusername/go-utils/strutil"

// Check if string is empty
empty := strutil.IsEmpty(nil) // true

s := "test"
empty = strutil.IsEmpty(&s) // false

// Convert to snake case
result := strutil.ToSnakeCase("HelloWorld") // "hello_world"
```

## Installation

```bash
go get github.com/yourusername/go-utils
```

## License

See [LICENSE](LICENSE) file for details.
