package utils

import "testing"

func TestValidatePasswordComplexity(t *testing.T) {
	tests := []struct {
		name      string
		password  string
		shouldErr bool
	}{
		{
			name:      "valid lower upper digit",
			password:  "Abcdef12",
			shouldErr: false,
		},
		{
			name:      "valid lower upper special",
			password:  "Abcdef!@",
			shouldErr: false,
		},
		{
			name:      "valid upper digit special",
			password:  "ABC123!@",
			shouldErr: false,
		},
		{
			name:      "valid lower digit special",
			password:  "abc123!@",
			shouldErr: false,
		},
		{
			name:      "invalid too short",
			password:  "Ab1!xyz",
			shouldErr: true,
		},
		{
			name:      "invalid only lower and digit",
			password:  "abcdef12",
			shouldErr: true,
		},
		{
			name:      "invalid only upper and digit",
			password:  "ABCDEF12",
			shouldErr: true,
		},
		{
			name:      "invalid only lower and special",
			password:  "abcdef!@",
			shouldErr: true,
		},
		{
			name:      "invalid only upper and special",
			password:  "ABCDEF!@",
			shouldErr: true,
		},
		{
			name:      "invalid only letters",
			password:  "Abcdefgh",
			shouldErr: true,
		},
		{
			name:      "invalid only digits and special",
			password:  "123456!@",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordComplexity(tt.password)
			if tt.shouldErr && err == nil {
				t.Fatalf("expected error, got nil for password %q", tt.password)
			}
			if !tt.shouldErr && err != nil {
				t.Fatalf("expected nil error, got %v for password %q", err, tt.password)
			}
		})
	}
}
