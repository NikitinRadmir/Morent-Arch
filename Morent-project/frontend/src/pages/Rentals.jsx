import React, { useCallback, useContext, useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import RentalCard from '../components/RentalCard';
import { AuthContext } from '../context/AuthContext';

const Rentals = () => {
    const { isAuthenticated, authRequest } = useContext(AuthContext);
    const [rentals, setRentals] = useState([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');

    const fetchRentals = useCallback(async () => {
        try {
            setLoading(true);
            setError('');
            const data = await authRequest('/rentals', { method: 'GET' });
            setRentals(Array.isArray(data) ? data : []);
        } catch (err) {
            setError(err.message || 'Failed to load rentals');
        } finally {
            setLoading(false);
        }
    }, [authRequest]);

    useEffect(() => {
        if (isAuthenticated) {
            fetchRentals();
        }
    }, [isAuthenticated, fetchRentals]);

    if (!isAuthenticated) {
        return (
            <div className="auth-page">
                <div className="auth-card">
                    <h2>My rentals</h2>
                    <p className="auth-subtitle">Please sign in to see your rental history.</p>
                    <Link className="auth-submit text-center d-block" to="/sign-in">Go to Sign in</Link>
                </div>
            </div>
        );
    }

    return (
        <div className="background-gray py-4 rentals-page">
            <div className="container">
                <div className="d-flex align-items-center justify-content-between mb-4">
                    <h2>My rentals</h2>
                    <button className="btn btn-outline-primary" onClick={fetchRentals} disabled={loading}>
                        Refresh
                    </button>
                </div>
                {error && <div className="alert alert-danger">{error}</div>}
                {loading ? (
                    <p>Loading rentals...</p>
                ) : rentals.length === 0 ? (
                    <p>You have no rentals yet.</p>
                ) : (
                    <div className="rental-list">
                        {rentals.map((rental) => (
                            <RentalCard key={rental.id} rental={rental} />
                        ))}
                    </div>
                )}
            </div>
        </div>
    );
};

export default Rentals;

