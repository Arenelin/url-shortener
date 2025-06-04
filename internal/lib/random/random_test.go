package random

import (
	"github.com/go-playground/assert/v2"
	"strings"
	"testing"
)

func TestNewRandomString(t *testing.T) {
	testCases := []struct {
		name string
		size int
	}{
		{
			name: "size - 1",
			size: 1,
		},
		{
			name: "size - 5",
			size: 5,
		},
		{
			name: "size - 5 second_version",
			size: 5,
		},
		{
			name: "size - 25",
			size: 25,
		},
	}
	uniqueStr := ""

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			randStr := NewRandomString(tc.size)
			isNotUniqueRandStr := strings.Contains(uniqueStr, randStr)
			uniqueStr = uniqueStr + " " + randStr

			assert.Equal(t, tc.size, len(randStr))
			assert.Equal(t, isNotUniqueRandStr, false)
		})
	}
}
