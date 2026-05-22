import React, { useContext, useState } from 'react';
import { AuthContext } from '../context/AuthContext';
import { Link, useLocation, useNavigate } from 'react-router-dom';

const SignIn = () => {
    const navigate = useNavigate();
    const location = useLocation();
    const redirectTo = location.state?.from || '/';
    const { login } = useContext(AuthContext);
    const [form, setForm] = useState({ email: '', password: '' });
    const [status, setStatus] = useState({ type: '', message: '' });
    const [loading, setLoading] = useState(false);

    const handleChange = (e) => {
        const { name, value } = e.target;
        setForm((prev) => ({ ...prev, [name]: value }));
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        setLoading(true);
        setStatus({ type: '', message: '' });
        const result = await login(form);
        setLoading(false);

        if (result.success) {
            if (result.requiresEmailVerification) {
                navigate('/verify-email', { replace: true });
                return;
            }
            setStatus({ type: 'success', message: result.message });
            navigate(redirectTo);
        } else {
            setStatus({ type: 'error', message: result.message });
        }
    };

    return (
        <div className="auth-page">
            <div className="auth-card">
                <h2>Sign in</h2>
                <p className="auth-subtitle">Welcome back! Please enter your details.</p>
                {status.message && (
                    <div className={`auth-alert auth-alert--${status.type}`}>
                        {status.message}
                    </div>
                )}
                <form className="auth-form" onSubmit={handleSubmit}>
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
                            placeholder="Your password"
                            required
                        />
                    </label>
                    <button className="auth-submit" type="submit" disabled={loading}>
                        {loading ? 'Signing in...' : 'Sign in'}
                    </button>
                </form>
                <p className="auth-meta">
                    Don't have an account? <Link to="/sign-up">Create one</Link>
                </p>
            </div>
        </div>
    );
};

export default SignIn;

