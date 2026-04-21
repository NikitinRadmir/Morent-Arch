import React, { useContext, useEffect } from 'react';
import { Link } from 'react-router-dom';
import CarsSection from '../components/CarsSection';
import { AuthContext } from '../context/AuthContext';

const Favorites = () => {
    const { isAuthenticated, favorites, favoritesLoading, refreshFavorites } = useContext(AuthContext);

    useEffect(() => {
        if (isAuthenticated) {
            refreshFavorites();
        }
    }, [isAuthenticated, refreshFavorites]);

    if (!isAuthenticated) {
        return (
            <div className="auth-page">
                <div className="auth-card">
                    <h2>Favorites</h2>
                    <p className="auth-subtitle">Please sign in to view your saved cars.</p>
                    <Link className="auth-submit text-center d-block" to="/sign-in">Go to Sign in</Link>
                </div>
            </div>
        );
    }

    if (favoritesLoading) {
        return (
            <div className="background-gray py-4">
                <div className="container">
                    <h2 className="mb-4">My Favorites</h2>
                    <p>Loading favorites...</p>
                </div>
            </div>
        );
    }

    if (favorites.length === 0) {
        return (
            <div className="background-gray py-4">
                <div className="container">
                    <h2 className="mb-4">My Favorites</h2>
                    <p>You have not added any cars to favorites yet.</p>
                </div>
            </div>
        );
    }

    return (
        <div className="background-gray">
            <CarsSection cars={favorites} title="My Favorites" showViewAll={false} col3 />
        </div>
    );
};

export default Favorites;

