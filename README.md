# jsonrepair

A Go library that repairs malformed JSON and JSON-adjacent text into valid JSON

## Features

- **Missing syntax repair**: Adds missing quotes, commas, colons, and closing brackets when the intent is clear.
- **Quote normalization**: Converts single quotes and smart quotes into valid JSON string quotes.
- **Comment removal**: Removes JavaScript-style line and block comments.
- **Wrapper cleanup**: Strips Markdown code fences, JSONP callbacks, MongoDB wrappers, and ellipsis placeholders.
- **Literal conversion**: Converts Python-style `None`, `True`, and `False` into JSON literals.
- **String recovery**: Repairs escaped strings, concatenated strings, and common truncated string input.
- **NDJSON repair**: Converts newline-delimited JSON values into one JSON array.

## Installation

```bash
go get github.com/kaptinlin/jsonrepair
```

Requires **Go 1.26+**.

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    "github.com/kaptinlin/jsonrepair"
)

func main() {
    repaired, err := jsonrepair.Repair("{name: 'John'}")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(repaired)
}
```

Output:

```json
{"name": "John"}
```

## API

```go
func Repair(text string) (string, error)
```

`Repair` returns the repaired JSON document or a structured `*jsonrepair.Error` when the input cannot be repaired.

```go
repaired, err := jsonrepair.Repair(input)
if err != nil {
    var repairErr *jsonrepair.Error
    if errors.As(err, &repairErr) {
        fmt.Printf("%s at position %d\n", repairErr.Message, repairErr.Position)
    }
}
```

Sentinel errors support `errors.Is`:

- `ErrUnexpectedEnd`
- `ErrObjectKeyExpected`
- `ErrColonExpected`
- `ErrInvalidCharacter`
- `ErrUnexpectedCharacter`
- `ErrInvalidUnicode`

## Development

```bash
task test    # Run tests with race detection
task lint    # Run golangci-lint and go mod tidy check
task         # Run lint and test
```

## Contributing

Contributions are welcome. Open an issue to discuss substantial changes before submitting a pull request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgements

This library is a Go port of Jos de Jong's JavaScript `jsonrepair` library.
