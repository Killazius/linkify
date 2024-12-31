package random_test

import (
	"github.com/stretchr/testify/assert"
	"shorturl/internal/lib/random"
	"testing"
)

func TestNewRandomString(t *testing.T) {
	type testCase struct {
		name string
		size int
	}

	testCases := []testCase{
		{
			name: "size = 5",
			size: 5,
		},
		{
			name: "size = 20",
			size: 20,
		},
		{
			name: "size = 10",
			size: 10,
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			str1 := random.NewRandomString(testCase.size)
			str2 := random.NewRandomString(testCase.size)

			assert.Len(t, str1, testCase.size)
			assert.Len(t, str2, testCase.size)

			assert.NotEqual(t, str1, str2)
		})
	}
}
