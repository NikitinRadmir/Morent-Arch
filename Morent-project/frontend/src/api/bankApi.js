import { API_BASE_URL } from '../context/AuthContext';

const bankFetch = async (path, options = {}) => {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    method: options.method || 'GET',
    headers: {
      'Content-Type': 'application/json',
      ...(options.headers || {}),
    },
    credentials: 'include',
    body: options.body,
  });
  const text = await response.text();
  let data = null;
  if (text) {
    try {
      data = JSON.parse(text);
    } catch {
      data = { error: text };
    }
  }
  if (!response.ok) {
    const message = data?.error || data?.message || text || response.statusText;
    throw new Error(message);
  }
  return data;
};

export const bankApi = {
  register: (payload) =>
    bankFetch('/bank/auth/register', {
      method: 'POST',
      body: JSON.stringify(payload),
    }),

  login: (payload) =>
    bankFetch('/bank/auth/login', {
      method: 'POST',
      body: JSON.stringify(payload),
    }),

  logout: () =>
    bankFetch('/bank/auth/logout', {
      method: 'POST',
    }),

  profile: () => bankFetch('/bank/profile'),

  deposit: (amount) =>
    bankFetch('/bank/deposit', {
      method: 'POST',
      body: JSON.stringify({ amount: Number(amount) }),
    }),

  transfer: (recipientPhone, amount) =>
    bankFetch('/bank/transfer', {
      method: 'POST',
      body: JSON.stringify({
        recipientPhone,
        amount: Number(amount),
      }),
    }),

  transactions: (limit = 20) => bankFetch(`/bank/transactions?limit=${limit}`),
};
