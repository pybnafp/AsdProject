package sms

import "testing"

func TestIsValidMobile(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"13812345678", true},   // valid
		{"19912345678", true},   // valid
		{"12345678901", false},  // invalid: starts with 12
		{"1381234567", false},   // invalid: too short
		{"23812345678", false},  // invalid: starts with 2
		{"138123456789", false}, // invalid: too long
		{"", false},             // invalid: empty
		{"abcdefghijk", false},  // invalid: non-numeric
		{"13900000000", true},   // valid
		{"15012345678", true},   // valid
	}

	for _, test := range tests {
		result := IsValidMobile(test.input)
		if result != test.expected {
			t.Errorf("IsValidMobile(%q) = %v; want %v", test.input, result, test.expected)
		}
	}
}
