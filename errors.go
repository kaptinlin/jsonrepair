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

func newError(message string, position int, err error) *Error {
	return &Error{Message: message, Position: position, Err: err}
}

func newUnexpectedEndError(position int) *Error {
	return newError(ErrUnexpectedEnd.Error(), position, ErrUnexpectedEnd)
}

func newObjectKeyExpectedError(position int) *Error {
	return newError(ErrObjectKeyExpected.Error(), position, ErrObjectKeyExpected)
}

func newColonExpectedError(position int) *Error {
	return newError(ErrColonExpected.Error(), position, ErrColonExpected)
}

func newUnexpectedCharacterError(message string, position int) *Error {
	return newError(message, position, ErrUnexpectedCharacter)
}

func newInvalidUnicodeError(message string, position int) *Error {
	return newError(message, position, ErrInvalidUnicode)
}

func newInvalidCharacterError(message string, position int) *Error {
	return newError(message, position, ErrInvalidCharacter)
}
