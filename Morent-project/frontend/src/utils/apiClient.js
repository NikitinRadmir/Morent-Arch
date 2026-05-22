/**
 * Унифицированная обработка сетевых и HTTP-ошибок (500, 502, 503).
 */

const DEFAULT_MESSAGES = {
    network: 'Сервер недоступен. Проверьте подключение или запустите make up.',
    500: 'Внутренняя ошибка сервера. Попробуйте позже.',
    502: 'Сервис-прокси вернул ошибку. Попробуйте позже.',
    503: 'Сервис временно недоступен. Попробуйте через несколько минут.',
    504: 'Превышено время ожидания ответа сервера.',
};

export function messageForStatus(status) {
    if (status === 503) return DEFAULT_MESSAGES[503];
    if (status === 502) return DEFAULT_MESSAGES[502];
    if (status === 504) return DEFAULT_MESSAGES[504];
    if (status >= 500) return DEFAULT_MESSAGES[500];
    return null;
}

export async function parseResponseError(response) {
    try {
        const text = await response.text();
        try {
            const data = JSON.parse(text);
            if (data?.error) return data.error;
            if (data?.message) return data.message;
        } catch {
            if (text?.trim()) return text.trim();
        }
    } catch {
        // ignore
    }
    return messageForStatus(response.status) || response.statusText || 'Request failed';
}

/**
 * fetch с credentials; при сетевой ошибке или 5xx — понятное сообщение.
 */
export async function resilientFetch(url, options = {}) {
    let response;
    try {
        response = await fetch(url, { credentials: 'include', ...options });
    } catch {
        const err = new Error(DEFAULT_MESSAGES.network);
        err.code = 'network';
        throw err;
    }

    if (!response.ok) {
        const fallback = messageForStatus(response.status);
        const message = fallback || (await parseResponseError(response));
        const err = new Error(message);
        err.status = response.status;
        if (response.status === 503) err.code = 'service_unavailable';
        if (response.status >= 500) err.code = 'server_error';
        throw err;
    }
    return response;
}

export async function resilientJson(url, options = {}) {
    const response = await resilientFetch(url, options);
    if (response.status === 204) return null;
    return response.json();
}
