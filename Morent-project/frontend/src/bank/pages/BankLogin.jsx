import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import toast from 'react-hot-toast';
import BankSimpleNav from '../components/BankSimpleNav';
import PhoneInput from '../components/PhoneInput';
import { useBank } from '../context/BankContext';

const BankLogin = () => {
  const navigate = useNavigate();
  const { login } = useBank();
  const [phone, setPhone] = useState('');
  const [password, setPassword] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!phone.trim() || !password) {
      toast.error('Заполните телефон и пароль');
      return;
    }
    setSubmitting(true);
    try {
      await login(phone, password);
      toast.success('Добро пожаловать');
      navigate('/bank');
    } catch (err) {
      toast.error(err.message || 'Ошибка входа');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <>
      <BankSimpleNav
        extraLink={
          <Link to="/bank/register" className="mb-nav-extra-link">
            Регистрация
          </Link>
        }
      />
      <div className="mb-auth-wrapper">
        <div className="mb-auth-card">
          <h1 className="mb-auth-title">Вход</h1>
          <form onSubmit={handleSubmit}>
            <div className="mb-3">
              <label className="form-label mb-form-label" htmlFor="login-phone">
                Телефон
              </label>
              <PhoneInput id="login-phone" name="phone" value={phone} onChange={setPhone} />
            </div>
            <div className="mb-4">
              <label className="form-label mb-form-label" htmlFor="login-password">
                Пароль
              </label>
              <input
                type="password"
                className="form-control mb-form-control"
                id="login-password"
                name="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                autoComplete="current-password"
              />
            </div>
            <button type="submit" className="btn mb-btn-submit" disabled={submitting}>
              {submitting ? 'Вход…' : 'Войти'}
            </button>
          </form>
          <p className="text-center mt-3 mb-0 mb-link-muted">
            Нет аккаунта? <Link to="/bank/register">Зарегистрироваться</Link>
          </p>
        </div>
      </div>
    </>
  );
};

export default BankLogin;
