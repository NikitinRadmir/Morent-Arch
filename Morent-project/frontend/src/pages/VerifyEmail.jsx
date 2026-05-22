import React, { useContext, useEffect, useRef, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { AuthContext } from '../context/AuthContext';

const CODE_LENGTH = 6;

const VerifyEmail = () => {
    const navigate = useNavigate();
    const {
        user,
        pendingVerificationEmail,
        verifyEmail,
        resendVerificationEmail,
    } = useContext(AuthContext);
    const [digits, setDigits] = useState(Array(CODE_LENGTH).fill(''));
    const [status, setStatus] = useState({ type: '', message: '' });
    const [loading, setLoading] = useState(false);
    const [resending, setResending] = useState(false);
    const inputsRef = useRef([]);

    const displayEmail = pendingVerificationEmail || user?.email;

    useEffect(() => {
        if (!displayEmail) {
            navigate('/sign-up', { replace: true });
            return;
        }
        if (user?.emailVerified) {
            navigate('/', { replace: true });
        }
    }, [displayEmail, user, navigate]);

    const code = digits.join('');

    const handleDigitChange = (index, value) => {
        const digit = value.replace(/\D/g, '').slice(-1);
        const next = [...digits];
        next[index] = digit;
        setDigits(next);
        if (digit && index < CODE_LENGTH - 1) {
            inputsRef.current[index + 1]?.focus();
        }
    };

    const handleKeyDown = (index, e) => {
        if (e.key === 'Backspace' && !digits[index] && index > 0) {
            inputsRef.current[index - 1]?.focus();
        }
    };

    const handlePaste = (e) => {
        const pasted = e.clipboardData.getData('text').replace(/\D/g, '').slice(0, CODE_LENGTH);
        if (!pasted) return;
        e.preventDefault();
        const next = Array(CODE_LENGTH).fill('');
        for (let i = 0; i < pasted.length; i += 1) {
            next[i] = pasted[i];
        }
        setDigits(next);
        const focusIndex = Math.min(pasted.length, CODE_LENGTH - 1);
        inputsRef.current[focusIndex]?.focus();
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        if (code.length !== CODE_LENGTH) {
            setStatus({ type: 'error', message: 'Введите 6 цифр кода из письма' });
            return;
        }
        setLoading(true);
        setStatus({ type: '', message: '' });
        const result = await verifyEmail(code);
        setLoading(false);
        if (result.success) {
            setStatus({ type: 'success', message: 'Email подтверждён. Добро пожаловать!' });
            setTimeout(() => navigate('/'), 800);
        } else {
            setStatus({ type: 'error', message: result.message });
        }
    };

    const handleResend = async () => {
        setResending(true);
        setStatus({ type: '', message: '' });
        const result = await resendVerificationEmail();
        setResending(false);
        setStatus({
            type: result.success ? 'success' : 'error',
            message: result.message,
        });
    };

    if (!displayEmail) {
        return null;
    }

    return (
        <div className="auth-page">
            <div className="auth-card">
                <h2>Подтверждение email</h2>
                <p className="auth-subtitle">
                    Мы отправили 6-значный код на <strong>{displayEmail}</strong>
                </p>
                <p className="auth-subtitle" style={{ marginTop: '-12px', fontSize: '14px' }}>
                    Вход в Morent будет доступен после подтверждения.
                </p>
                {status.message && (
                    <div className={`auth-alert auth-alert--${status.type}`}>{status.message}</div>
                )}
                <form className="auth-form" onSubmit={handleSubmit}>
                    <div className="verify-code-row" onPaste={handlePaste}>
                        {digits.map((digit, index) => (
                            <input
                                key={index}
                                ref={(el) => { inputsRef.current[index] = el; }}
                                type="text"
                                inputMode="numeric"
                                maxLength={1}
                                className="verify-code-input"
                                value={digit}
                                onChange={(e) => handleDigitChange(index, e.target.value)}
                                onKeyDown={(e) => handleKeyDown(index, e)}
                                aria-label={`Цифра ${index + 1}`}
                            />
                        ))}
                    </div>
                    <button className="auth-submit" type="submit" disabled={loading || code.length !== CODE_LENGTH}>
                        {loading ? 'Проверка…' : 'Подтвердить и войти'}
                    </button>
                </form>
                <p className="auth-meta">
                    Не пришло письмо?{' '}
                    <button
                        type="button"
                        className="auth-link-button"
                        onClick={handleResend}
                        disabled={resending}
                    >
                        {resending ? 'Отправка…' : 'Отправить код снова'}
                    </button>
                </p>
                <p className="auth-meta">
                    <Link to="/sign-in">Уже подтверждали? Войти</Link>
                </p>
            </div>
        </div>
    );
};

export default VerifyEmail;
