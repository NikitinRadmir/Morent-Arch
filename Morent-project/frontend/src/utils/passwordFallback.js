const SPECIAL = '!@#$%^&*()-_=+[]{}?';

export const PASSWORD_REQUIREMENTS = [
    { id: 'length', label: 'От 10 до 14 символов' },
    { id: 'lower', label: 'Строчная буква (a-z)' },
    { id: 'upper', label: 'Заглавная буква (A-Z)' },
    { id: 'digit', label: 'Цифра (0-9)' },
    { id: 'special', label: 'Спецсимвол (!@#$%^&*…)' },
];

/** Локальная проверка, если generator-service недоступен. */
export function validatePasswordFallback(password) {
    const length = password.length;
    const hasLower = /[a-z]/.test(password);
    const hasUpper = /[A-Z]/.test(password);
    const hasDigit = /\d/.test(password);
    const hasSpecial = [...password].some((ch) => SPECIAL.includes(ch));
    const lengthOk = length >= 10 && length <= 14;

    const errors = [];
    if (!lengthOk) errors.push('длина пароля должна быть от 10 до 14 символов');
    if (!hasLower) errors.push('нужна хотя бы одна строчная буква (a-z)');
    if (!hasUpper) errors.push('нужна хотя бы одна заглавная буква (A-Z)');
    if (!hasDigit) errors.push('нужна хотя бы одна цифра');
    if (!hasSpecial) errors.push('нужен хотя бы один спецсимвол');

    return {
        valid: lengthOk && hasLower && hasUpper && hasDigit && hasSpecial,
        errors,
        hasLower,
        hasUpper,
        hasDigit,
        hasSpecial,
        lengthOk,
        length,
        fallback: true,
    };
}

export function requirementStatus(result) {
    if (!result) {
        return {
            length: false,
            lower: false,
            upper: false,
            digit: false,
            special: false,
        };
    }
    return {
        length: result.lengthOk,
        lower: result.hasLower,
        upper: result.hasUpper,
        digit: result.hasDigit,
        special: result.hasSpecial,
    };
}
