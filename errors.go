package jsonrepair

import (
	"errors"
	"fmt"
)

// Predefined error variables for use with errors.Is()
var (
	ErrUnexpectedEnd       = errors.New("unexpected end of json string")
	ErrObjectKeyExpected   = errors.New("object key expected")
	ErrColonExpected       = errors.New("colon expected")
	ErrInvalidCharacter    = errors.New("invalid character")
	ErrUnexpectedCharacter = errors.New("unexpected character")
	ErrInvalidUnicode      = errors.New("invalid unicode character")
)

// JSONRepairError represents a structured JSON repair error.
// It provides the error message, position, and optional underlying error
type JSONRepairError struct {
	Message  string
	Position int
	Err      error // optional underlying error
}

// Error implements the error interface
func (e *JSONRepairError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s at position %d: %v", e.Message, e.Position, e.Err)
	}
	return fmt.Sprintf("%s at position %d", e.Message, e.Position)
}

// Unwrap allows JSONRepairError to support errors.Is / errors.As
func (e *JSONRepairError) Unwrap() error {
	return e.Err
}

// newJSONRepairError creates a new JSONRepairError with optional error wrapping
// Usage:
//
//	newJSONRepairError("Unexpected character", 42)
//	newJSONRepairError("Invalid unicode character", 13, ErrInvalidUnicode)
//	newJSONRepairError("Unexpected character", 42, ErrUnexpectedCharacter)
func newJSONRepairError(message string, position int, err ...error) *JSONRepairError {
	var inner error
	if len(err) > 0 {
		inner = err[0]
	}
	return &JSONRepairError{Message: message, Position: position, Err: inner}
}

// Convenience functions for creating specific error types with predefined errors wrapped
func newUnexpectedEndError(position int) *JSONRepairError {
	return newJSONRepairError("Unexpected end of json string", position, ErrUnexpectedEnd)
}

func newObjectKeyExpectedError(position int) *JSONRepairError {
	return newJSONRepairError("Object key expected", position, ErrObjectKeyExpected)
}

func newColonExpectedError(position int) *JSONRepairError {
	return newJSONRepairError("Colon expected", position, ErrColonExpected)
}

func newUnexpectedCharacterError(message string, position int) *JSONRepairError {
	return newJSONRepairError(message, position, ErrUnexpectedCharacter)
}

func newInvalidUnicodeError(message string, position int) *JSONRepairError {
	return newJSONRepairError(message, position, ErrInvalidUnicode)
}

func newInvalidCharacterError(message string, position int) *JSONRepairError {
	return newJSONRepairError(message, position, ErrInvalidCharacter)
}
