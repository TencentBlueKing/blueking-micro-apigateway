package stringx

import (
	"testing"
)

func TestRandString(t *testing.T) {
	length := 10
	result := RandString(length)
	if len(result) != length {
		t.Errorf("Expected length %d, but got %d", length, len(result))
	}
	for _, char := range result {
		if !contains(letterBytes, byte(char)) {
			t.Errorf("Unexpected character %c in result", char)
		}
	}
}

func contains(s string, c byte) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return true
		}
	}
	return false
}
