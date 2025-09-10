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

// Error represents a structured JSON repair error.
// It provides the error message, position, and optional underlying error
type Error struct {
	Message  string
	Position int
	Err      error // optional underlying error
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s at position %d: %v", e.Message, e.Position, e.Err)
	}
	return fmt.Sprintf("%s at position %d", e.Message, e.Position)
}

// Unwrap allows Error to support errors.Is / errors.As
func (e *Error) Unwrap() error {
	return e.Err
}

// newJSONRepairError creates a new Error with optional error wrapping
// Usage:
//
//	newJSONRepairError("Unexpected character", 42)
//	newJSONRepairError("Invalid unicode character", 13, ErrInvalidUnicode)
//	newJSONRepairError("Unexpected character", 42, ErrUnexpectedCharacter)
func newJSONRepairError(message string, position int, err ...error) *Error {
	var inner error
	if len(err) > 0 {
		inner = err[0]
	}
	return &Error{Message: message, Position: position, Err: inner}
}

// Convenience functions for creating specific error types with predefined errors wrapped
func newUnexpectedEndError(position int) *Error {
	return newJSONRepairError("Unexpected end of json string", position, ErrUnexpectedEnd)
}

func newObjectKeyExpectedError(position int) *Error {
	return newJSONRepairError("Object key expected", position, ErrObjectKeyExpected)
}

func newColonExpectedError(position int) *Error {
	return newJSONRepairError("Colon expected", position, ErrColonExpected)
}

func newUnexpectedCharacterError(message string, position int) *Error {
	return newJSONRepairError(message, position, ErrUnexpectedCharacter)
}

func newInvalidUnicodeError(message string, position int) *Error {
	return newJSONRepairError(message, position, ErrInvalidUnicode)
}

func newInvalidCharacterError(message string, position int) *Error {
	return newJSONRepairError(message, position, ErrInvalidCharacter)
}
