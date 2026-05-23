import { API_BASE_URL } from '../context/AuthContext';
import { resilientFetch } from '../utils/apiClient';

export async function graphqlRequest(query, variables = {}, options = {}) {
    const resp = await resilientFetch(`${API_BASE_URL}/graphql`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            ...(options.headers || {}),
        },
        body: JSON.stringify({ query, variables }),
    });
    const text = await resp.text();
    let json = {};
    try {
        json = text ? JSON.parse(text) : {};
    } catch {
        throw new Error('Сервер вернул некорректный GraphQL-ответ');
    }
    if (!resp.ok || json.errors) {
        const msg = json?.errors?.[0]?.message || resp.statusText || 'GraphQL request failed';
        throw new Error(msg);
    }
    return json.data;
}

