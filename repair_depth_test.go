package jsonrepair

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type nestingCase struct {
	name         string
	open         string
	close        string
	empty        string
	openerOffset int
	unwrap       bool
}

var nestingCases = []*nestingCase{
	{name: "arrays", open: "[", close: "]", empty: "[]"},
	{name: "objects", open: `{"a":`, close: "}", empty: "{}"},
	{name: "function calls", open: "f(", close: ")", openerOffset: 1, unwrap: true},
}

func (tc *nestingCase) build(depth int, closed bool) (input, want string) {
	prefix := strings.Repeat(tc.open, depth)
	suffix := strings.Repeat(tc.close, depth)
	input = prefix + "0"
	if closed {
		input += suffix
	}
	want = prefix + "0" + suffix
	if tc.unwrap {
		want = "0"
	}
	return input, want
}

func (tc *nestingCase) buildEmpty(depth int) string {
	return strings.Repeat(tc.open, depth-1) + tc.empty + strings.Repeat(tc.close, depth-1)
}

func (tc *nestingCase) exceededPosition(depth int) int {
	return len([]rune(tc.open))*depth + tc.openerOffset
}

func TestRepairNestingDepthBoundary(t *testing.T) {
	t.Parallel()

	for _, tc := range nestingCases {
		t.Run(tc.name, func(t *testing.T) {
			input, want := tc.build(maxNestingDepth, true)
			got, err := Repair(input)
			require.NoError(t, err)
			assert.Equal(t, want, got)

			for _, closed := range []bool{true, false} {
				input, _ = tc.build(maxNestingDepth+1, closed)
				requireMaxDepthError(t, input, tc.exceededPosition(maxNestingDepth))
			}
		})
	}
}

func TestRepairEmptyNestingDepthBoundary(t *testing.T) {
	t.Parallel()

	for _, tc := range nestingCases {
		if tc.empty == "" {
			continue
		}
		t.Run(tc.name, func(t *testing.T) {
			input := tc.buildEmpty(maxNestingDepth)
			got, err := Repair(input)
			require.NoError(t, err)
			assert.Equal(t, input, got)

			requireMaxDepthError(
				t,
				tc.buildEmpty(maxNestingDepth+1),
				tc.exceededPosition(maxNestingDepth),
			)
		})
	}
}

func TestRepairNestingDepthInNDJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		prefix string
	}{
		{name: "first entry"},
		{name: "second entry", prefix: "0\n"},
		{name: "later entry", prefix: "0\n1\n"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			input := tc.prefix + strings.Repeat("[", maxNestingDepth+1)
			requireMaxDepthError(t, input, len([]rune(tc.prefix))+maxNestingDepth)
		})
	}
}

func requireMaxDepthError(t *testing.T, input string, position int) {
	t.Helper()

	got, err := Repair(input)
	require.ErrorIs(t, err, ErrMaxDepthExceeded)
	assert.Empty(t, got)

	var repairErr *Error
	require.ErrorAs(t, err, &repairErr)
	assert.Equal(t, ErrMaxDepthExceeded.Error(), repairErr.Message)
	assert.Equal(t, position, repairErr.Position)
	assert.Same(t, ErrMaxDepthExceeded, repairErr.Err)
}
