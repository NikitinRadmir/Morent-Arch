import React, { createContext, useCallback, useEffect, useMemo, useState } from 'react';

export const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:1488';
const AUTH_USER_KEY = 'morent_auth_user';
const AUTH_TOKEN_KEY = 'morent_auth_token';

export const AuthContext = createContext({
    user: null,
    token: null,
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

const persistState = (user, token) => {
    if (user) {
        localStorage.setItem(AUTH_USER_KEY, JSON.stringify(user));
    } else {
        localStorage.removeItem(AUTH_USER_KEY);
    }

    if (token) {
        localStorage.setItem(AUTH_TOKEN_KEY, token);
    } else {
        localStorage.removeItem(AUTH_TOKEN_KEY);
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

export const AuthProvider = ({ children }) => {
    const [user, setUser] = useState(() => {
        const stored = localStorage.getItem(AUTH_USER_KEY);
        return stored ? JSON.parse(stored) : null;
    });
    const [token, setToken] = useState(() => localStorage.getItem(AUTH_TOKEN_KEY));
    const [favorites, setFavorites] = useState([]);
    const [favoritesLoading, setFavoritesLoading] = useState(false);

    useEffect(() => {
        persistState(user, token);
    }, [user, token]);

    const sendRequest = useCallback(async (path, payload) => {
        const response = await fetch(`${API_BASE_URL}${path}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(payload),
        });
        if (!response.ok) {
            const message = await parseError(response);
            throw new Error(message);
        }

        return response.json();
    }, []);

    const authRequest = useCallback(async (path, options = {}) => {
        if (!token) {
            throw new Error('Not authenticated');
        }
        const response = await fetch(`${API_BASE_URL}${path}`, {
            method: options.method || 'GET',
            headers: {
                'Content-Type': 'application/json',
                Authorization: `Bearer ${token}`,
                ...(options.headers || {}),
            },
            body: options.body,
        });
        if (!response.ok) {
            const message = await parseError(response);
            throw new Error(message);
        }
        if (response.status === 204) {
            return null;
        }
        const text = await response.text();
        return text ? JSON.parse(text) : null;
    }, [token]);

    const loadFavorites = useCallback(async () => {
        if (!token) {
            setFavorites([]);
            return;
        }
        try {
            setFavoritesLoading(true);
            const data = await authRequest('/favorites');
            setFavorites(Array.isArray(data) ? data : []);
        } catch (error) {
            console.error('Failed to load favorites', error);
        } finally {
            setFavoritesLoading(false);
        }
    }, [token, authRequest]);

    const register = useCallback(async ({ name, email, password }) => {
        try {
            const data = await sendRequest('/auth/register', { name, email, password });
            setUser(data.user);
            setToken(data.token);
            setFavorites([]);
            return { success: true, message: 'Registration successful' };
        } catch (error) {
            return { success: false, message: error.message };
        }
    }, [sendRequest]);

    const login = useCallback(async ({ email, password }) => {
        try {
            const data = await sendRequest('/auth/login', { email, password });
            setUser(data.user);
            setToken(data.token);
            await loadFavorites();
            return { success: true, message: 'Welcome back!' };
        } catch (error) {
            return { success: false, message: error.message };
        }
    }, [sendRequest, loadFavorites]);

    const logout = useCallback(() => {
        setUser(null);
        setToken(null);
        setFavorites([]);
    }, []);

    const fetchProfile = useCallback(async () => {
        if (!token) return null;
        const data = await authRequest('/auth/profile');
        setUser(data);
        return data;
    }, [authRequest, token]);

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
        await authRequest('/favorites', {
            method: 'POST',
            body: JSON.stringify({ carId }),
        });
        await loadFavorites();
    }, [authRequest, loadFavorites]);

    const removeFavorite = useCallback(async (carId) => {
        await authRequest(`/favorites/${carId}`, {
            method: 'DELETE',
        });
        await loadFavorites();
    }, [authRequest, loadFavorites]);

    const isFavorite = useCallback((carId) => {
        return favorites.some((car) => car.id === carId);
    }, [favorites]);

    const toggleFavorite = useCallback(async (carId) => {
        if (!token) {
            throw new Error('Not authenticated');
        }
        if (isFavorite(carId)) {
            await removeFavorite(carId);
        } else {
            await addFavorite(carId);
        }
    }, [token, isFavorite, addFavorite, removeFavorite]);

    useEffect(() => {
        if (token) {
            loadFavorites();
            // Если роль отсутствует, обновляем профиль для получения актуальных данных
            if (user && !user.role && !user.Role) {
                fetchProfile().catch(console.error);
            }
        } else {
            setFavorites([]);
        }
    }, [token, loadFavorites, user, fetchProfile]);

    const value = useMemo(() => ({
        user,
        token,
        favorites,
        favoritesLoading,
        isAuthenticated: Boolean(user && token),
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
    }), [user, token, favorites, favoritesLoading, register, login, logout, loadFavorites, addFavorite, removeFavorite, toggleFavorite, isFavorite, authRequest, fetchProfile, updateProfile, changePassword]);

    return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

