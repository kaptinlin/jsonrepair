# Golang JSONRepair Library

Easily repair invalid JSON documents with the Golang JSONRepair Library. This library is a direct port of the popular [jsonrepair JavaScript library](https://github.com/josdejong/jsonrepair), designed to address common issues found in JSON data. Leveraging the performance benefits of Go, it maintains compatibility and reliability with the original JavaScript library. It is particularly useful for optimizing JSON content generated by language models (LLMs).

## Features

The `jsonrepair` library can automatically fix the following JSON issues:

- **Add missing quotes around keys**: Ensures all keys are properly quoted.
- **Add missing escape characters**: Adds necessary escape characters where needed.
- **Add missing commas**: Inserts missing commas between elements.
- **Add missing closing brackets**: Closes any unclosed brackets.
- **Repair truncated JSON**: Completes truncated JSON data.
- **Replace single quotes with double quotes**: Converts single quotes to double quotes.
- **Replace special quote characters**: Converts characters like `“...”` to standard double quotes.
- **Replace special white space characters**: Converts special whitespace characters to regular spaces.
- **Replace Python constants**: Converts `None`, `True`, `False` to `null`, `true`, `false`.
- **Strip trailing commas**: Removes any trailing commas.
- **Strip comments**: Eliminates comments such as `/* ... */` and `// ...`.
- **Strip fenced code blocks**: Removes markdown fenced code blocks like `` ```json`` and `` ``` ``.
- **Strip ellipsis**: Removes ellipsis in arrays and objects, e.g., `[1, 2, 3, ...]`.
- **Strip JSONP notation**: Removes JSONP callbacks, e.g., `callback({ ... })`.
- **Strip escape characters**: Removes escape characters from strings, e.g., `{\"stringified\": \"content\"}`.
- **Strip MongoDB data types**: Converts types like `NumberLong(2)` and `ISODate("2012-12-19T06:01:17.171Z")` to standard JSON.
- **Concatenate strings**: Merges strings split across lines, e.g., `"long text" + "more text on next line"`.
- **Convert newline-delimited JSON**: Encloses newline-delimited JSON in an array to make it valid, for example:

    ```json
    { "id": 1, "name": "John" }
    { "id": 2, "name": "Sarah" }
    ```

## Install

Install the library using `go get`:

```sh
go get github.com/kaptinlin/jsonrepair
```

## Usage

### Basic Usage

Use the `JSONRepair` function to repair a JSON string:

```go
package main

import (
    "fmt"
    "log"

    "github.com/kaptinlin/jsonrepair"
)

func main() {
    // The following is invalid JSON: it consists of JSON contents copied from
    // a JavaScript code base, where the keys are missing double quotes,
    // and strings are using single quotes:
    json := "{name: 'John'}"

    repaired, err := jsonrepair.JSONRepair(json)
    if err != nil {
        log.Fatalf("Failed to repair JSON: %v", err)
    }

    fmt.Println(repaired) // '{"name": "John"}'
}
```

## API

### JSONRepair Function

```go
// JSONRepair attempts to repair the given JSON string and returns the repaired version.
// It returns an error if an issue is encountered which could not be solved.
func JSONRepair(text string) (string, error)
```

## How to Contribute

Contributions to the `jsonrepair` package are welcome. If you'd like to contribute, please follow the [contribution guidelines](CONTRIBUTING.md).

## License

Released under the MIT license. See the [LICENSE](LICENSE) file for details.

## Acknowledgements

This library is a Go port of the JavaScript library `jsonrepair` by [Jos de Jong](https://github.com/josdejong). The original logic and behavior have been closely followed to ensure compatibility and reliability. Special thanks to the original author for creating such a useful tool.
