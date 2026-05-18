import { API_BASE_URL } from '../context/AuthContext';

/**
 * Запросы к Morent API (прокси к generator-service).
 * Пароль при валидации только в POST body — не в URL и не в query.
 */

async function readResponseBody(response) {
    const text = await response.text();
    if (!text) return {};
    try {
        return JSON.parse(text);
    } catch {
        return { error: text.trim() || 'Некорректный ответ сервера' };
    }
}

export async function generatePassword() {
    try {
        const response = await fetch(`${API_BASE_URL}/auth/password/generate`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
            cache: 'no-store',
        });
        const data = await readResponseBody(response);

        if (response.status === 503 || data.unavailable) {
            return { ok: false, unavailable: true };
        }
        if (!response.ok) {
            return {
                ok: false,
                error: data.error || data.message || 'Не удалось сгенерировать пароль',
            };
        }
        if (!data.password) {
            return { ok: false, error: 'Сервер не вернул пароль' };
        }
        return { ok: true, password: data.password };
    } catch {
        return { ok: false, unavailable: true };
    }
}

export async function validatePassword(password) {
    try {
        const response = await fetch(`${API_BASE_URL}/auth/password/validate`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
            cache: 'no-store',
            body: JSON.stringify({ password }),
        });
        const data = await readResponseBody(response);

        if (response.status === 503 || data.unavailable) {
            return { ok: false, unavailable: true };
        }
        if (!response.ok) {
            return { ok: false, error: data.error || 'Ошибка проверки пароля' };
        }
        return { ok: true, result: data };
    } catch {
        return { ok: false, unavailable: true };
    }
}
