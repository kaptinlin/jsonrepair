package jsonrepair

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRepairDeeplyNestedDoesNotOverflow ensures that pathologically deep input
// is rejected with an error instead of exhausting the goroutine stack and
// crashing the program. Objects, arrays and function-call values are parsed
// recursively, so an unbounded nesting depth is a denial-of-service vector.
func TestRepairDeeplyNestedDoesNotOverflow(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"arrays":         strings.Repeat("[", maxRecursionDepth+100),
		"objects":        strings.Repeat(`{"a":`, maxRecursionDepth+100),
		"function calls": strings.Repeat("f(", maxRecursionDepth+100),
		// A very large input that, without the depth limit, overflows the stack.
		"huge arrays": strings.Repeat("[", 2_000_000) + strings.Repeat("]", 2_000_000),
	}

	for name, input := range cases {
		input := input
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := Repair(input)
			require.Error(t, err)
			assert.True(t, errors.Is(err, ErrMaxDepthExceeded),
				"expected ErrMaxDepthExceeded, got %v", err)
		})
	}
}

// TestRepairDeepButValidNestingSucceeds ensures the depth limit is generous
// enough that realistic deep documents still parse.
func TestRepairDeepButValidNestingSucceeds(t *testing.T) {
	t.Parallel()

	const depth = 1000
	input := strings.Repeat("[", depth) + strings.Repeat("]", depth)
	out, err := Repair(input)
	require.NoError(t, err)
	assert.Equal(t, input, out)
}
