# go-utils

A lightweight Go utilities library providing helpful string manipulation functions and asynchronous utilities.

## Features

### String Utilities (`strutil`)

- **`IsEmpty(s *string) bool`** - Check if a string pointer is nil or empty
  - Returns `true` if the string is `nil` or has zero length

- **`ToSnakeCase(str string) string`** - Convert strings to snake_case format
  - Converts camelCase and PascalCase to snake_case
  - Example: `HelloWorld` → `hello_world`

### Future Utilities (`future`)

- **`StartFuture[T any](ctx context.Context, fn func(context.Context) (T, error)) *Future[T]`** - Launch a function in a goroutine and return a Future
  - Handles context cancellation and panics gracefully
  - Non-blocking send to avoid goroutine leaks

- **`Future[T].Await(ctx context.Context) (T, error)`** - Block until the Future completes or context is done
  - Only the first successful call to `Await` on a given `Future` will receive the result; subsequent calls will block indefinitely because the underlying value has already been consumed

- **`Future[T].Try() (T, error, bool)`** - Non-blocking check for completion
  - Only the first successful call to `Try` on a given `Future` will receive the result; subsequent calls will return `(_, _, false)` because the underlying value has already been consumed

- **`All[T any](ctx context.Context, futures []*Future[T]) ([]T, error)`** - Wait for all futures to complete

- **`Any[T any](ctx context.Context, futures []*Future[T]) (T, error, int)`** - Return the first completed future

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

```go
import "github.com/yourusername/go-utils/future"

// Start a future
fut := future.StartFuture(context.Background(), func(ctx context.Context) (string, error) {
    time.Sleep(time.Second)
    return "done", nil
})

// Await the result
result, err := fut.Await(context.Background())
if err != nil {
    log.Fatal(err)
}
fmt.Println(result) // "done"
```

## Installation

```bash
go get github.com/yourusername/go-utils
```

## License

See [LICENSE](LICENSE) file for details.
