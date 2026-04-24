package generators

import (
	"testing"
)

func TestGeneratePasswordByMask(t *testing.T) {
	tests := []struct {
		name    string
		mask    string
		wantErr bool
	}{
		{"valid mask", "LLLdddss", false},
		{"empty mask", "", true},
		{"mixed mask", "Ll-dd", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			password, err := GeneratePasswordByMask(tt.mask)
			if (err != nil) != tt.wantErr {
				t.Errorf("GeneratePasswordByMask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(password) != len(tt.mask) {
				t.Errorf("GeneratePasswordByMask() length = %v, want %v", len(password), len(tt.mask))
			}
		})
	}
}

func TestGenerateRandomMask(t *testing.T) {
	tests := []struct {
		name    string
		length  int
		wantErr bool
	}{
		{"valid length 10", 10, false},
		{"valid length 14", 14, false},
		{"too short", 9, true},
		{"too long", 15, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mask, err := GenerateRandomMask(tt.length)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateRandomMask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(mask) != tt.length {
				t.Errorf("GenerateRandomMask() length = %v, want %v", len(mask), tt.length)
			}
		})
	}
}

func TestGeneratePasswordAuto(t *testing.T) {
	password, mask, err := GeneratePasswordAuto()
	if err != nil {
		t.Errorf("GeneratePasswordAuto() error = %v", err)
		return
	}
	if len(password) < 10 || len(password) > 14 {
		t.Errorf("GeneratePasswordAuto() length = %v, want between 10 and 14", len(password))
	}
	if len(password) != len(mask) {
		t.Errorf("GeneratePasswordAuto() password length %v != mask length %v", len(password), len(mask))
	}
}
