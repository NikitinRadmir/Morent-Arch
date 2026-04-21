import React, { useContext, useState } from 'react';
import { AuthContext } from '../context/AuthContext';
import { Link, useNavigate } from 'react-router-dom';

const initialState = {
    name: '',
    email: '',
    password: '',
    confirmPassword: '',
};

const SignUp = () => {
    const navigate = useNavigate();
    const { register } = useContext(AuthContext);
    const [form, setForm] = useState(initialState);
    const [status, setStatus] = useState({ type: '', message: '' });
    const [loading, setLoading] = useState(false);

    const handleChange = (e) => {
        const { name, value } = e.target;
        setForm((prev) => ({ ...prev, [name]: value }));
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        if (form.password !== form.confirmPassword) {
            setStatus({ type: 'error', message: 'Passwords do not match' });
            return;
        }

        setLoading(true);
        setStatus({ type: '', message: '' });
        const result = await register({
            name: form.name,
            email: form.email,
            password: form.password,
        });
        setLoading(false);

        if (result.success) {
            setStatus({ type: 'success', message: result.message });
            navigate('/');
        } else {
            setStatus({ type: 'error', message: result.message });
        }
    };

    return (
        <div className="auth-page">
            <div className="auth-card">
                <h2>Create account</h2>
                <p className="auth-subtitle">Register to manage your rentals faster.</p>
                {status.message && (
                    <div className={`auth-alert auth-alert--${status.type}`}>
                        {status.message}
                    </div>
                )}
                <form className="auth-form" onSubmit={handleSubmit}>
                    <label>
                        Full name
                        <input
                            type="text"
                            name="name"
                            value={form.name}
                            onChange={handleChange}
                            placeholder="John Carter"
                            required
                        />
                    </label>
                    <label>
                        Email
                        <input
                            type="email"
                            name="email"
                            value={form.email}
                            onChange={handleChange}
                            placeholder="name@example.com"
                            required
                        />
                    </label>
                    <label>
                        Password
                        <input
                            type="password"
                            name="password"
                            value={form.password}
                            onChange={handleChange}
                            placeholder="Create password"
                            required
                            minLength={6}
                        />
                    </label>
                    <label>
                        Confirm password
                        <input
                            type="password"
                            name="confirmPassword"
                            value={form.confirmPassword}
                            onChange={handleChange}
                            placeholder="Repeat password"
                            required
                            minLength={6}
                        />
                    </label>
                    <button className="auth-submit" type="submit" disabled={loading}>
                        {loading ? 'Creating account...' : 'Sign up'}
                    </button>
                </form>
                <p className="auth-meta">
                    Already have an account? <Link to="/sign-in">Sign in</Link>
                </p>
            </div>
        </div>
    );
};

export default SignUp;

