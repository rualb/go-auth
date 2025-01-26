package utilcrypto

import "testing"

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "Valid password",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "Empty password",
			password: "",
			wantErr:  false, // No error should occur, but returns empty string
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HashPassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.password != "" && len(got) == 0 {
				t.Errorf("HashPassword() got = %v, want non-empty string", got)
			}
		})
	}
}

func TestCompareHashAndPassword(t *testing.T) {
	// First, we hash a password to use for comparison
	hash, err := HashPassword("password123")
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	tests := []struct {
		name     string
		hash     string
		password string
		want     bool
	}{
		{
			name:     "Correct password",
			hash:     hash,
			password: "password123",
			want:     true,
		},
		{
			name:     "Incorrect password",
			hash:     hash,
			password: "wrongpassword",
			want:     false,
		},
		{
			name:     "Empty hash",
			hash:     "",
			password: "password123",
			want:     false,
		},
		{
			name:     "Empty password",
			hash:     hash,
			password: "",
			want:     false,
		},
		{
			name:     "Both empty",
			hash:     "",
			password: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CompareHashAndPassword(tt.hash, tt.password)
			if got != tt.want {
				t.Errorf("CompareHashAndPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}
