import React, { createContext, useCallback, useEffect, useMemo, useState } from 'react';
import { bankApi } from '../api/bankApi';

export const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:1488';
const AUTH_USER_KEY = 'morent_auth_user';
const storage = window.sessionStorage;

export const AuthContext = createContext({
    user: null,
    isAuthenticated: false,
    favorites: [],
    favoritesLoading: false,
    register: async () => {},
    login: async () => {},
    logout: () => {},
    refreshFavorites: async () => {},
    addFavorite: async () => {},
    removeFavorite: async () => {},
    toggleFavorite: async () => {},
    isFavorite: () => false,
    authRequest: async () => {},
});

const persistState = (user) => {
    if (user) {
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
        return stored ? JSON.parse(stored) : null;
    });
    const [favorites, setFavorites] = useState([]);
    const [favoritesLoading, setFavoritesLoading] = useState(false);

    useEffect(() => {
        persistState(user);
    }, [user]);

    const clearAuthState = useCallback(() => {
        setUser(null);
        setFavorites([]);
    }, []);

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
            }
            throw new Error(message);
        }
        if (response.status === 204) {
            return null;
        }
        const text = await response.text();
        return text ? JSON.parse(text) : null;
    }, [clearAuthState]);

    const loadFavorites = useCallback(async () => {
        if (!user) {
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
            const data = await sendRequest('/auth/register', { name, email, password });
            setUser(data.user);
            setFavorites([]);
            try {
                await bankApi.syncSession();
            } catch (bankErr) {
                console.warn('Bank card provisioning failed', bankErr);
            }
            return { success: true, message: 'Registration successful' };
        } catch (error) {
            return { success: false, message: error.message };
        }
    }, [sendRequest]);

    const login = useCallback(async ({ email, password }) => {
        try {
            const data = await sendRequest('/auth/login', { email, password });
            setUser(data.user);
            await loadFavorites();
            try {
                await bankApi.syncSession();
            } catch (bankErr) {
                console.warn('Bank session sync failed', bankErr);
            }
            return { success: true, message: 'Welcome back!' };
        } catch (error) {
            return { success: false, message: error.message };
        }
    }, [sendRequest, loadFavorites]);

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
        setUser(data);
        return data;
    }, [authRequest]);

    const updateProfile = useCallback(async (payload) => {
        const data = await authRequest('/auth/profile', {
            method: 'PUT',
            body: JSON.stringify(payload),
        });
        setUser(data);
        return data;
    }, [authRequest]);

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
        if (!user) {
            throw new Error('Not authenticated');
        }
        if (isFavorite(carId)) {
            await removeFavorite(carId);
        } else {
            await addFavorite(carId);
        }
    }, [user, isFavorite, addFavorite, removeFavorite]);

    useEffect(() => {
        if (user) {
            loadFavorites();
            // Если роль отсутствует, обновляем профиль для получения актуальных данных
            if (user && !user.role && !user.Role) {
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
        favorites,
        favoritesLoading,
        isAuthenticated: Boolean(user),
        register,
        login,
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
    }), [user, favorites, favoritesLoading, register, login, logout, loadFavorites, addFavorite, removeFavorite, toggleFavorite, isFavorite, authRequest, fetchProfile, updateProfile, changePassword]);

    return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

