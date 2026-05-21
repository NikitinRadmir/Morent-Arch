import React, { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react';
import toast from 'react-hot-toast';
import { AuthContext } from '../../context/AuthContext';
import { bankApi } from '../../api/bankApi';

const BankContext = createContext(null);

export const BankProvider = ({ children }) => {
  const { isAuthenticated, user } = useContext(AuthContext);
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

  const bootstrap = useCallback(async () => {
    if (!isAuthenticated) {
      setProfile(null);
      setLoading(false);
      return;
    }
    setLoading(true);
    try {
      const sessionProfile = await bankApi.syncSession();
      if (sessionProfile?.cardNumber) {
        setProfile(sessionProfile);
      } else {
        await refreshProfile();
      }
    } catch (err) {
      setProfile(null);
      toast.error(err.message || 'Не удалось открыть банковский счёт');
    } finally {
      setLoading(false);
    }
  }, [isAuthenticated, refreshProfile]);

  useEffect(() => {
    bootstrap();
  }, [bootstrap, user?.id]);

  const deposit = useCallback(async (amount) => {
    const data = await bankApi.deposit(amount);
    setProfile(data.profile);
    toast.success('Счёт пополнен');
    return data;
  }, []);

  const transfer = useCallback(async (recipientCardNumber, amount) => {
    const data = await bankApi.transfer(recipientCardNumber, amount);
    setProfile(data.profile);
    toast.success('Перевод выполнен');
    return data;
  }, []);

  const value = useMemo(
    () => ({
      profile,
      loading,
      isAuthenticated: Boolean(profile),
      deposit,
      transfer,
      refreshProfile,
    }),
    [profile, loading, deposit, transfer, refreshProfile],
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
