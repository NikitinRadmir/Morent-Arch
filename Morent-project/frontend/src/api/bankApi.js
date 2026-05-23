import { API_BASE_URL } from '../context/AuthContext';
import { resilientJson } from '../utils/apiClient';

const bankFetch = async (path, options = {}) => {
  return resilientJson(`${API_BASE_URL}${path}`, {
    method: options.method || 'GET',
    headers: {
      'Content-Type': 'application/json',
      ...(options.headers || {}),
    },
    body: options.body,
  });
};

export const bankApi = {
  syncSession: () =>
    bankFetch('/bank/session', {
      method: 'POST',
    }),

  profile: () => bankFetch('/bank/profile'),

  deposit: (amount) =>
    bankFetch('/bank/deposit', {
      method: 'POST',
      body: JSON.stringify({ amount: Number(amount) }),
    }),

  transfer: (recipientCardNumber, amount) =>
    bankFetch('/bank/transfer', {
      method: 'POST',
      body: JSON.stringify({
        recipientCardNumber: String(recipientCardNumber).replace(/\D/g, ''),
        amount: Number(amount),
      }),
    }),

  transactions: (limit = 20) => bankFetch(`/bank/transactions?limit=${limit}`),
};
