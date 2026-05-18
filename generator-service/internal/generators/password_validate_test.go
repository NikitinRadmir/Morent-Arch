package generators

import "testing"

func TestValidatePassword(t *testing.T) {
	ok := ValidatePassword("Ab1!xxxxxx")
	if !ok.Valid {
		t.Error("expected valid password")
	}
	bad := ValidatePassword("short")
	if bad.Valid {
		t.Error("expected invalid short password")
	}
}
