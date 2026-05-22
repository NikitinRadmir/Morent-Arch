import React, { createContext, useCallback, useEffect, useMemo, useState } from 'react';
import { bankApi } from '../api/bankApi';

export const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:1488';
const AUTH_USER_KEY = 'morent_auth_user';
const PENDING_VERIFY_EMAIL_KEY = 'morent_pending_verify_email';
const storage = window.sessionStorage;

const isEmailNotVerifiedError = (message) => {
    const text = (message || '').toLowerCase();
    return text.includes('email is not verified')
        || text.includes('подтвердите email');
};

export const AuthContext = createContext({
    user: null,
    isAuthenticated: false,
    pendingVerificationEmail: null,
    favorites: [],
    favoritesLoading: false,
    register: async () => {},
    login: async () => {},
    verifyEmail: async () => {},
    resendVerificationEmail: async () => {},
    logout: () => {},
    refreshFavorites: async () => {},
    addFavorite: async () => {},
    removeFavorite: async () => {},
    toggleFavorite: async () => {},
    isFavorite: () => false,
    authRequest: async () => {},
});

const persistUser = (user) => {
    if (user?.emailVerified) {
        storage.setItem(AUTH_USER_KEY, JSON.stringify(user));
    } else {
        storage.removeItem(AUTH_USER_KEY);
    }
};

const parseError = async (response) => {
    try {
        const text = await response.text();
        try {
            const data = JSON.parse(text);
            if (data && data.error) {
                return data.error;
            }
            if (typeof data === 'string') {
                return data;
            }
        } catch {
            if (text) {
                return text;
            }
        }
    } catch {
        // ignore
    }
    return response.statusText || 'Request failed';
};

const normalizeFavorites = (data) => {
    if (!Array.isArray(data)) return [];
    return data
        .map((item) => ({ ...item, id: Number(item.id) }))
        .filter((item) => Number.isFinite(item.id));
};

export const AuthProvider = ({ children }) => {
    const [user, setUser] = useState(() => {
        const stored = storage.getItem(AUTH_USER_KEY);
        if (!stored) return null;
        try {
            const parsed = JSON.parse(stored);
            return parsed?.emailVerified ? parsed : null;
        } catch {
            return null;
        }
    });
    const [pendingVerificationEmail, setPendingVerificationEmail] = useState(
        () => storage.getItem(PENDING_VERIFY_EMAIL_KEY) || null,
    );
    const [favorites, setFavorites] = useState([]);
    const [favoritesLoading, setFavoritesLoading] = useState(false);

    const setPendingEmail = useCallback((email) => {
        const normalized = email ? String(email).trim().toLowerCase() : '';
        if (normalized) {
            storage.setItem(PENDING_VERIFY_EMAIL_KEY, normalized);
            setPendingVerificationEmail(normalized);
        } else {
            storage.removeItem(PENDING_VERIFY_EMAIL_KEY);
            setPendingVerificationEmail(null);
        }
    }, []);

    const clearAuthState = useCallback(() => {
        setUser(null);
        setFavorites([]);
        storage.removeItem(AUTH_USER_KEY);
    }, []);

    useEffect(() => {
        persistUser(user);
    }, [user]);

    const sendRequest = useCallback(async (path, payload) => {
        const response = await fetch(`${API_BASE_URL}${path}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            credentials: 'include',
            body: JSON.stringify(payload),
        });
        if (!response.ok) {
            const message = await parseError(response);
            throw new Error(message);
        }

        return response.json();
    }, []);

    const authRequest = useCallback(async (path, options = {}) => {
        const response = await fetch(`${API_BASE_URL}${path}`, {
            method: options.method || 'GET',
            headers: {
                'Content-Type': 'application/json',
                ...(options.headers || {}),
            },
            credentials: 'include',
            body: options.body,
        });
        if (!response.ok) {
            const message = await parseError(response);
            if (response.status === 401) {
                clearAuthState();
                setPendingEmail(null);
            }
            throw new Error(message);
        }
        if (response.status === 204) {
            return null;
        }
        const text = await response.text();
        return text ? JSON.parse(text) : null;
    }, [clearAuthState, setPendingEmail]);

    const loadFavorites = useCallback(async () => {
        if (!user?.emailVerified) {
            setFavorites([]);
            return;
        }
        try {
            setFavoritesLoading(true);
            const data = await authRequest('/favorites');
            setFavorites(normalizeFavorites(data));
        } catch (error) {
            console.error('Failed to load favorites', error);
            setFavorites([]);
        } finally {
            setFavoritesLoading(false);
        }
    }, [user, authRequest]);

    const register = useCallback(async ({ name, email, password }) => {
        try {
            clearAuthState();
            setPendingEmail(null);
            await fetch(`${API_BASE_URL}/auth/logout`, {
                method: 'POST',
                credentials: 'include',
            }).catch(() => {});

            const data = await sendRequest('/auth/register', { name, email, password });
            const normalizedEmail = (data.user?.email || email || '').trim().toLowerCase();
            setPendingEmail(normalizedEmail);
            return {
                success: true,
                message: 'Проверьте почту и введите код подтверждения',
                requiresEmailVerification: Boolean(data.requiresEmailVerification ?? true),
            };
        } catch (error) {
            return { success: false, message: error.message };
        }
    }, [sendRequest, clearAuthState, setPendingEmail]);

    const login = useCallback(async ({ email, password }) => {
        try {
            const data = await sendRequest('/auth/login', { email, password });
            const normalizedEmail = (data.user?.email || email || '').trim().toLowerCase();

            if (data.requiresEmailVerification) {
                clearAuthState();
                setPendingEmail(normalizedEmail);
                return {
                    success: true,
                    requiresEmailVerification: true,
                    message: 'Аккаунт не подтверждён. Введите код из письма — мы отправили его снова.',
                };
            }

            setPendingEmail(null);
            setUser(data.user);
            await loadFavorites();
            try {
                await bankApi.syncSession();
            } catch (bankErr) {
                console.warn('Bank session sync failed', bankErr);
            }
            return { success: true, message: 'Welcome back!' };
        } catch (error) {
            if (isEmailNotVerifiedError(error.message)) {
                clearAuthState();
                setPendingEmail(email);
                return {
                    success: true,
                    requiresEmailVerification: true,
                    message: 'Подтвердите email — введите код из письма',
                };
            }
            return { success: false, message: error.message };
        }
    }, [sendRequest, loadFavorites, clearAuthState, setPendingEmail]);

    const logout = useCallback(async () => {
        try {
            await fetch(`${API_BASE_URL}/auth/logout`, {
                method: 'POST',
                credentials: 'include',
            });
        } catch (error) {
            console.warn('Logout request failed', error);
        } finally {
            clearAuthState();
        }
    }, [clearAuthState]);

    const fetchProfile = useCallback(async () => {
        const data = await authRequest('/auth/profile');
        if (data?.emailVerified) {
            setUser(data);
            setPendingEmail(null);
        } else {
            clearAuthState();
            if (data?.email) {
                setPendingEmail(data.email);
            }
        }
        return data;
    }, [authRequest, clearAuthState, setPendingEmail]);

    const updateProfile = useCallback(async (payload) => {
        const data = await authRequest('/auth/profile', {
            method: 'PUT',
            body: JSON.stringify(payload),
        });
        setUser(data);
        return data;
    }, [authRequest]);

    const verifyEmail = useCallback(async (code) => {
        const email = pendingVerificationEmail || storage.getItem(PENDING_VERIFY_EMAIL_KEY);
        if (!email) {
            return { success: false, message: 'Сначала зарегистрируйтесь или войдите' };
        }
        try {
            const data = await sendRequest('/auth/verify-email', { email, code });
            setPendingEmail(null);
            setUser(data.user);
            try {
                await bankApi.syncSession();
            } catch (bankErr) {
                console.warn('Bank card provisioning failed', bankErr);
            }
            return { success: true, message: 'Email подтверждён' };
        } catch (error) {
            return { success: false, message: error.message };
        }
    }, [sendRequest, pendingVerificationEmail, setPendingEmail]);

    const resendVerificationEmail = useCallback(async () => {
        const email = pendingVerificationEmail || storage.getItem(PENDING_VERIFY_EMAIL_KEY);
        if (!email) {
            return { success: false, message: 'Укажите email при регистрации' };
        }
        try {
            await sendRequest('/auth/verify-email/resend', { email });
            return { success: true, message: 'Код отправлен повторно' };
        } catch (error) {
            return { success: false, message: error.message };
        }
    }, [sendRequest, pendingVerificationEmail]);

    const changePassword = useCallback(async ({ oldPassword, newPassword }) => {
        await authRequest('/auth/password', {
            method: 'PUT',
            body: JSON.stringify({ oldPassword, newPassword }),
        });
        return true;
    }, [authRequest]);

    const addFavorite = useCallback(async (carId) => {
        try {
            await authRequest('/favorites', {
                method: 'POST',
                body: JSON.stringify({ carId: Number(carId) }),
            });
        } finally {
            await loadFavorites();
        }
    }, [authRequest, loadFavorites]);

    const removeFavorite = useCallback(async (carId) => {
        try {
            await authRequest(`/favorites/${Number(carId)}`, {
                method: 'DELETE',
            });
        } finally {
            await loadFavorites();
        }
    }, [authRequest, loadFavorites]);

    const isFavorite = useCallback((carId) => {
        const normalized = Number(carId);
        return favorites.some((car) => Number(car.id) === normalized);
    }, [favorites]);

    const toggleFavorite = useCallback(async (carId) => {
        if (!user?.emailVerified) {
            throw new Error('Not authenticated');
        }
        if (isFavorite(carId)) {
            await removeFavorite(carId);
        } else {
            await addFavorite(carId);
        }
    }, [user, isFavorite, addFavorite, removeFavorite]);

    useEffect(() => {
        if (user?.emailVerified) {
            loadFavorites();
            if (!user.role && !user.Role) {
                fetchProfile().catch(console.error);
            }
        } else {
            setFavorites([]);
        }
    }, [loadFavorites, user, fetchProfile]);

    useEffect(() => {
        if (!user) {
            fetchProfile().catch(() => {});
        }
    }, [user, fetchProfile]);

    const value = useMemo(() => ({
        user,
        pendingVerificationEmail,
        favorites,
        favoritesLoading,
        isAuthenticated: Boolean(user?.emailVerified),
        register,
        login,
        verifyEmail,
        resendVerificationEmail,
        logout,
        refreshFavorites: loadFavorites,
        addFavorite,
        removeFavorite,
        toggleFavorite,
        isFavorite,
        authRequest,
        fetchProfile,
        updateProfile,
        changePassword,
    }), [
        user,
        pendingVerificationEmail,
        favorites,
        favoritesLoading,
        register,
        login,
        verifyEmail,
        resendVerificationEmail,
        logout,
        loadFavorites,
        addFavorite,
        removeFavorite,
        toggleFavorite,
        isFavorite,
        authRequest,
        fetchProfile,
        updateProfile,
        changePassword,
    ]);

    return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};
