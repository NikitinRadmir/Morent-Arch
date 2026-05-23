import { API_BASE_URL } from '../context/AuthContext';
import { resilientJson } from '../utils/apiClient';

/**
 * Запросы к Morent API (прокси к generator-service).
 * Пароль при валидации только в POST body — не в URL и не в query.
 */

export async function generatePassword() {
    try {
        const data = await resilientJson(`${API_BASE_URL}/auth/password/generate`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            cache: 'no-store',
        });

        if (data?.unavailable) {
            return { ok: false, unavailable: true };
        }
        if (!data?.password) {
            return { ok: false, error: 'Сервер не вернул пароль' };
        }
        return { ok: true, password: data.password };
    } catch (error) {
        return { ok: false, unavailable: true, error: error.message };
    }
}

function passwordServiceUnavailable(error) {
    return error?.code === 'network' || error?.status === 502 || error?.status === 503 || error?.status === 504 || error?.status >= 500;
}

export async function validatePassword(password) {
    try {
        const data = await resilientJson(`${API_BASE_URL}/auth/password/validate`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            cache: 'no-store',
            body: JSON.stringify({ password }),
        });

        if (data?.unavailable) {
            return { ok: false, unavailable: true };
        }
        return { ok: true, result: data };
    } catch (error) {
        if (passwordServiceUnavailable(error)) {
            return { ok: false, unavailable: true, error: error.message };
        }
        return { ok: false, error: error.message || 'Ошибка проверки пароля' };
    }
}
