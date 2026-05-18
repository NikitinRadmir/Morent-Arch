import React, { useContext, useState } from 'react';
import { AuthContext } from '../context/AuthContext';
import { Link, useNavigate } from 'react-router-dom';
import NewPasswordField from '../components/NewPasswordField';
import PasswordInput from '../components/PasswordInput';
import { usePasswordField } from '../hooks/usePasswordField';

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
    const passwordField = usePasswordField(form.password);

    const handleChange = (e) => {
        const { name, value } = e.target;
        setForm((prev) => ({ ...prev, [name]: value }));
        if (name === 'password') {
            passwordField.resetGeneratorState();
        }
    };

    const handleSubmit = async (e) => {
        e.preventDefault();

        if (form.password !== form.confirmPassword) {
            setStatus({ type: 'error', message: 'Пароли не совпадают' });
            return;
        }

        if (!passwordField.isValid) {
            setStatus({
                type: 'error',
                message: 'Пароль не соответствует требованиям безопасности',
            });
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

    const passwordsMatch =
        form.confirmPassword.length > 0 && form.password === form.confirmPassword;
    const canSubmit =
        !loading &&
        !passwordField.checking &&
        passwordField.isValid &&
        passwordsMatch &&
        form.name.trim() &&
        form.email.trim();

    return (
        <div className="auth-page">
            <div className="auth-card">
                <h2>Создать аккаунт</h2>
                <p className="auth-subtitle">Регистрация для быстрого управления арендой.</p>
                {status.message && (
                    <div className={`auth-alert auth-alert--${status.type}`}>{status.message}</div>
                )}
                <form className="auth-form" onSubmit={handleSubmit} autoComplete="off">
                    <label>
                        Имя
                        <input
                            type="text"
                            name="name"
                            value={form.name}
                            onChange={handleChange}
                            placeholder="Иван Иванов"
                            required
                            autoComplete="name"
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
                            autoComplete="email"
                        />
                    </label>
                    <NewPasswordField
                        label="Пароль"
                        value={form.password}
                        onChange={(value) => {
                            setForm((prev) => ({ ...prev, password: value }));
                            passwordField.resetGeneratorState();
                        }}
                        onGenerated={(value) => {
                            setForm((prev) => ({ ...prev, password: value, confirmPassword: value }));
                        }}
                        disabled={loading}
                        requirementsId="signup-password-requirements"
                        placeholder="Создайте пароль"
                        passwordField={passwordField}
                    />
                    <label>
                        Подтвердите пароль
                        <PasswordInput
                            name="confirmPassword"
                            value={form.confirmPassword}
                            onChange={handleChange}
                            placeholder="Повторите пароль"
                            required
                            autoComplete="new-password"
                            aria-invalid={form.confirmPassword.length > 0 && !passwordsMatch}
                        />
                        {form.confirmPassword.length > 0 && !passwordsMatch && (
                            <p className="password-hint password-hint--error">Пароли не совпадают</p>
                        )}
                    </label>
                    <button className="auth-submit" type="submit" disabled={!canSubmit}>
                        {loading ? 'Создание…' : 'Зарегистрироваться'}
                    </button>
                </form>
                <p className="auth-meta">
                    Уже есть аккаунт? <Link to="/sign-in">Войти</Link>
                </p>
            </div>
        </div>
    );
};

export default SignUp;
