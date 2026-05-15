import React, { createContext, useCallback, useContext, useMemo, useState } from 'react';

const STORAGE_KEY = 'morent_bank_session';

const defaultSession = {
  isLoggedIn: false,
  phone: '',
  userName: 'Александр Петров',
  role: 'Клиент',
  balance: 125430,
};

function loadSession() {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return defaultSession;
    return { ...defaultSession, ...JSON.parse(raw) };
  } catch {
    return defaultSession;
  }
}

function saveSession(session) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(session));
}

const BankContext = createContext(null);

export const BankProvider = ({ children }) => {
  const [session, setSession] = useState(loadSession);

  const persist = useCallback((next) => {
    setSession(next);
    saveSession(next);
  }, []);

  const login = useCallback(
    (phone) => {
      persist({
        ...session,
        isLoggedIn: true,
        phone,
      });
    },
    [persist, session],
  );

  const register = useCallback(
    (phone) => {
      persist({
        ...session,
        isLoggedIn: true,
        phone,
        userName: 'Новый клиент',
      });
    },
    [persist, session],
  );

  const logout = useCallback(() => {
    persist({
      ...defaultSession,
      balance: session.balance,
    });
  }, [persist, session.balance]);

  const deposit = useCallback(
    (amount) => {
      const value = Number(amount);
      if (!Number.isFinite(value) || value <= 0) {
        return { ok: false, message: 'Укажите корректную сумму' };
      }
      const next = { ...session, balance: session.balance + value };
      persist(next);
      return { ok: true, message: `Счёт пополнен на ${value.toLocaleString('ru-RU')} ₽` };
    },
    [persist, session],
  );

  const transfer = useCallback(
    (amount) => {
      const value = Number(amount);
      if (!Number.isFinite(value) || value <= 0) {
        return { ok: false, message: 'Укажите корректную сумму' };
      }
      if (value > session.balance) {
        return { ok: false, message: 'Недостаточно средств на счёте' };
      }
      const next = { ...session, balance: session.balance - value };
      persist(next);
      return { ok: true, message: `Перевод ${value.toLocaleString('ru-RU')} ₽ выполнен` };
    },
    [persist, session],
  );

  const value = useMemo(
    () => ({
      ...session,
      login,
      register,
      logout,
      deposit,
      transfer,
      formatBalance: (amount = session.balance) =>
        `${amount.toLocaleString('ru-RU')} ₽`,
    }),
    [session, login, register, logout, deposit, transfer],
  );

  return <BankContext.Provider value={value}>{children}</BankContext.Provider>;
};

export const useBank = () => {
  const ctx = useContext(BankContext);
  if (!ctx) {
    throw new Error('useBank must be used within BankProvider');
  }
  return ctx;
};
