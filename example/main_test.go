package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/kaptinlin/jsonrepair"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errWriteFailed = errors.New("write failed")

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errWriteFailed
}

func TestRunWritesRepairedJSON(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer

	err := run(&output, "{name: 'John'}")

	require.NoError(t, err)
	assert.Equal(t, "{\"name\": \"John\"}\n", output.String())
}

func TestRunReturnsWriterError(t *testing.T) {
	t.Parallel()

	err := run(failingWriter{}, "{name: 'John'}")

	require.ErrorIs(t, err, errWriteFailed)
}

func TestMainWritesRepairedJSON(t *testing.T) {
	// os.Stdout is process-wide.
	originalStdout := os.Stdout
	reader, writer, err := os.Pipe()
	require.NoError(t, err)
	t.Cleanup(func() { os.Stdout = originalStdout })
	os.Stdout = writer

	main()

	require.NoError(t, writer.Close())
	output, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.NoError(t, reader.Close())
	assert.Equal(t, "{\"name\": \"John\"}\n", string(output))
}

func TestRunReturnsRepairError(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer

	err := run(&output, `{"a":2}foo`)

	require.ErrorIs(t, err, jsonrepair.ErrUnexpectedCharacter)
	assert.Empty(t, output.String())

	repairErr, ok := errors.AsType[*jsonrepair.Error](err)
	require.True(t, ok)
	assert.Equal(t, `unexpected character "f"`, repairErr.Message)
	assert.Equal(t, 7, repairErr.Position)
	assert.Same(t, jsonrepair.ErrUnexpectedCharacter, repairErr.Err)
}
