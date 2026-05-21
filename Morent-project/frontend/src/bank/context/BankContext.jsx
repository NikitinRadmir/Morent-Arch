import React, { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react';
import toast from 'react-hot-toast';
import { bankApi } from '../../api/bankApi';

const BankContext = createContext(null);

export const BankProvider = ({ children }) => {
  const [profile, setProfile] = useState(null);
  const [loading, setLoading] = useState(true);

  const refreshProfile = useCallback(async () => {
    try {
      const data = await bankApi.profile();
      setProfile(data);
      return data;
    } catch {
      setProfile(null);
      return null;
    }
  }, []);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      setLoading(true);
      try {
        await bankApi.profile();
        if (!cancelled) {
          await refreshProfile();
        }
      } catch {
        if (!cancelled) {
          setProfile(null);
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [refreshProfile]);

  const login = useCallback(async (phone, password) => {
    const data = await bankApi.login({ phone, password });
    setProfile(data.profile);
    return data;
  }, []);

  const register = useCallback(async (phone, password, displayName = '') => {
    const data = await bankApi.register({ phone, password, displayName });
    setProfile(data.profile);
    return data;
  }, []);

  const logout = useCallback(async () => {
    try {
      await bankApi.logout();
    } finally {
      setProfile(null);
    }
  }, []);

  const deposit = useCallback(async (amount) => {
    const data = await bankApi.deposit(amount);
    setProfile(data.profile);
    toast.success('Счёт пополнен');
    return data;
  }, []);

  const transfer = useCallback(async (recipientPhone, amount) => {
    const data = await bankApi.transfer(recipientPhone, amount);
    setProfile(data.profile);
    toast.success('Перевод выполнен');
    return data;
  }, []);

  const value = useMemo(
    () => ({
      profile,
      loading,
      isAuthenticated: Boolean(profile),
      login,
      register,
      logout,
      deposit,
      transfer,
      refreshProfile,
    }),
    [profile, loading, login, register, logout, deposit, transfer, refreshProfile],
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
