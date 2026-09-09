package utils

import "testing"

func TestIsValidClusterIdentifier(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		want       bool
	}{
		{"internal id", "abcdef0123456789abcdef0123456789", true},
		{"external id", "00000000-0000-4000-8000-000000000000", true},
		{"cluster display name", "my-test-cluster-01", true},
		{"management cluster name", "hs-mc-abc123xyz", true},

		{"empty", "", false},
		{"quote", "x' or external_id like '654321", false},
		{"backslash and quote", `x\' or 1=1 --`, false},
		{"percent", "6543%", false},
		{"underscore", "654_21", false},
		{"whitespace", "654321 or id = '1'", false},
		{"newline", "654321\nid", false},
		{"free-text placeholder", "<no value>", false},
		{"trailing punctuation", "654321,", false},
		{"over length limit", string(make([]byte, 65)), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidClusterIdentifier(tt.identifier); got != tt.want {
				t.Errorf("IsValidClusterIdentifier(%q) = %v, want %v", tt.identifier, got, tt.want)
			}
		})
	}
}
