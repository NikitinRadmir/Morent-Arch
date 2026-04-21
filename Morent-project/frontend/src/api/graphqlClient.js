import { API_BASE_URL } from '../context/AuthContext';

export async function graphqlRequest(query, variables = {}, options = {}) {
    const resp = await fetch(`${API_BASE_URL}/graphql`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            ...(options.headers || {}),
        },
        body: JSON.stringify({ query, variables }),
    });
    const json = await resp.json();
    if (!resp.ok || json.errors) {
        const msg = json?.errors?.[0]?.message || resp.statusText || 'GraphQL request failed';
        throw new Error(msg);
    }
    return json.data;
}

