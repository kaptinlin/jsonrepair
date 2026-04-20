package jsonrepair

import (
	"errors"
	"fmt"
)

var (
	// ErrUnexpectedEnd reports that the input ended before the JSON value was complete.
	ErrUnexpectedEnd = errors.New("unexpected end of json string")
	// ErrObjectKeyExpected reports that an object key could not be parsed.
	ErrObjectKeyExpected = errors.New("object key expected")
	// ErrColonExpected reports that an object key is not followed by a colon.
	ErrColonExpected = errors.New("colon expected")
	// ErrInvalidCharacter reports an invalid character inside a JSON string.
	ErrInvalidCharacter = errors.New("invalid character")
	// ErrUnexpectedCharacter reports a character that cannot be repaired in context.
	ErrUnexpectedCharacter = errors.New("unexpected character")
	// ErrInvalidUnicode reports an invalid Unicode escape sequence.
	ErrInvalidUnicode = errors.New("invalid unicode character")
)

// Error reports a non-repairable parse error and its position.
type Error struct {
	// Message describes the repair failure.
	Message string
	// Position is the rune offset where parsing failed.
	Position int
	// Err is the underlying sentinel for errors.Is and errors.As.
	Err error
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
