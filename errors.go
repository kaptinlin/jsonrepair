package jsonrepair

import (
	"errors"
	"fmt"
)

// Predefined error variables for use with errors.Is().
var (
	ErrUnexpectedEnd       = errors.New("unexpected end of json string")
	ErrObjectKeyExpected   = errors.New("object key expected")
	ErrColonExpected       = errors.New("colon expected")
	ErrInvalidCharacter    = errors.New("invalid character")
	ErrUnexpectedCharacter = errors.New("unexpected character")
	ErrInvalidUnicode      = errors.New("invalid unicode character")
)

// Error represents a structured JSON repair error with position information.
type Error struct {
	Message  string
	Position int
	Err      error // underlying error for errors.Is/As support
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s at position %d: %v", e.Message, e.Position, e.Err)
	}
	return fmt.Sprintf("%s at position %d", e.Message, e.Position)
}

// Unwrap returns the underlying error for errors.Is/As support.
func (e *Error) Unwrap() error {
	return e.Err
}

// newError creates a new Error wrapping the given sentinel error.
func newError(message string, position int, err error) *Error {
	return &Error{Message: message, Position: position, Err: err}
}

// newUnexpectedEndError creates an error for unexpected end of JSON input.
func newUnexpectedEndError(position int) *Error {
	return newError("unexpected end of json string", position, ErrUnexpectedEnd)
}

// newObjectKeyExpectedError creates an error for a missing object key.
func newObjectKeyExpectedError(position int) *Error {
	return newError("object key expected", position, ErrObjectKeyExpected)
}

// newColonExpectedError creates an error for a missing colon separator.
func newColonExpectedError(position int) *Error {
	return newError("colon expected", position, ErrColonExpected)
}

// newUnexpectedCharacterError creates an error for an unexpected character.
func newUnexpectedCharacterError(message string, position int) *Error {
	return newError(message, position, ErrUnexpectedCharacter)
}

// newInvalidUnicodeError creates an error for an invalid unicode escape sequence.
func newInvalidUnicodeError(message string, position int) *Error {
	return newError(message, position, ErrInvalidUnicode)
}

// newInvalidCharacterError creates an error for an invalid character in a string.
func newInvalidCharacterError(message string, position int) *Error {
	return newError(message, position, ErrInvalidCharacter)
}
