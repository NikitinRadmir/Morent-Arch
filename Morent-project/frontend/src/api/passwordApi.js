import { API_BASE_URL } from '../context/AuthContext';

/**
 * Запросы к Morent API (прокси к generator-service).
 * Пароль при валидации только в POST body по HTTPS — не в URL и не в query.
 */
async function parseJsonResponse(response) {
    const text = await response.text();
    try {
        return text ? JSON.parse(text) : {};
    } catch {
        throw new Error('Некорректный ответ сервера');
    }
}

export async function generatePassword() {
    const response = await fetch(`${API_BASE_URL}/auth/password/generate`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        cache: 'no-store',
    });
    const data = await parseJsonResponse(response);
    if (response.status === 503) {
        return { ok: false, unavailable: true };
    }
    if (!response.ok) {
        return { ok: false, error: data.error || 'Не удалось сгенерировать пароль' };
    }
    return { ok: true, password: data.password };
}

export async function validatePassword(password) {
    const response = await fetch(`${API_BASE_URL}/auth/password/validate`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        cache: 'no-store',
        body: JSON.stringify({ password }),
    });
    const data = await parseJsonResponse(response);
    if (response.status === 503 || data.unavailable) {
        return { ok: false, unavailable: true };
    }
    if (!response.ok) {
        return { ok: false, error: data.error || 'Ошибка проверки пароля' };
    }
    return { ok: true, result: data };
}
