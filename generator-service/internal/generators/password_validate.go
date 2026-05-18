package generators

import (
	"strings"
	"unicode"
)

const (
	PasswordMinLen = 10
	PasswordMaxLen = 14
)

// PasswordValidationResult — результат проверки пароля по правилам генератора.
type PasswordValidationResult struct {
	Valid        bool     `json:"valid"`
	Errors       []string `json:"errors,omitempty"`
	HasLower     bool     `json:"hasLower"`
	HasUpper     bool     `json:"hasUpper"`
	HasDigit     bool     `json:"hasDigit"`
	HasSpecial   bool     `json:"hasSpecial"`
	LengthOK     bool     `json:"lengthOk"`
	Length       int      `json:"length"`
}

// ValidatePassword проверяет пароль: длина 10–14, строчные, заглавные, цифра, спецсимвол.
func ValidatePassword(password string) PasswordValidationResult {
	res := PasswordValidationResult{
		Length: len(password),
	}

	if res.Length < PasswordMinLen || res.Length > PasswordMaxLen {
		res.Errors = append(res.Errors, "длина пароля должна быть от 10 до 14 символов")
	} else {
		res.LengthOK = true
	}

	for _, r := range password {
		switch {
		case unicode.IsLower(r):
			res.HasLower = true
		case unicode.IsUpper(r):
			res.HasUpper = true
		case unicode.IsDigit(r):
			res.HasDigit = true
		}
		if strings.ContainsRune(specialSymbols, r) {
			res.HasSpecial = true
		}
	}

	if !res.HasLower {
		res.Errors = append(res.Errors, "нужна хотя бы одна строчная буква (a-z)")
	}
	if !res.HasUpper {
		res.Errors = append(res.Errors, "нужна хотя бы одна заглавная буква (A-Z)")
	}
	if !res.HasDigit {
		res.Errors = append(res.Errors, "нужна хотя бы одна цифра")
	}
	if !res.HasSpecial {
		res.Errors = append(res.Errors, "нужен хотя бы один спецсимвол (!@#$%^&*()-_=+[]{}?)")
	}

	res.Valid = res.LengthOK && res.HasLower && res.HasUpper && res.HasDigit && res.HasSpecial
	return res
}
