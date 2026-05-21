import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import toast from 'react-hot-toast';
import BankSimpleNav from '../components/BankSimpleNav';
import PhoneInput from '../components/PhoneInput';
import { useBank } from '../context/BankContext';

const BankRegister = () => {
  const navigate = useNavigate();
  const { register } = useBank();
  const [phone, setPhone] = useState('');
  const [password, setPassword] = useState('');
  const [passwordConfirm, setPasswordConfirm] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!phone.trim() || !password || !passwordConfirm) {
      toast.error('Заполните все поля');
      return;
    }
    if (password !== passwordConfirm) {
      toast.error('Пароли не совпадают');
      return;
    }
    setSubmitting(true);
    try {
      await register(phone, password);
      toast.success('Регистрация успешна');
      navigate('/bank');
    } catch (err) {
      toast.error(err.message || 'Ошибка регистрации');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <>
      <BankSimpleNav
        extraLink={
          <Link to="/bank/login" className="mb-nav-extra-link">
            Вход
          </Link>
        }
      />
      <div className="mb-auth-wrapper">
        <div className="mb-auth-card">
          <h1 className="mb-auth-title">Регистрация</h1>
          <form onSubmit={handleSubmit}>
            <div className="mb-3">
              <label className="form-label mb-form-label" htmlFor="reg-phone">
                Телефон
              </label>
              <PhoneInput id="reg-phone" name="phone" value={phone} onChange={setPhone} />
            </div>
            <div className="mb-3">
              <label className="form-label mb-form-label" htmlFor="reg-password">
                Пароль
              </label>
              <input
                type="password"
                className="form-control mb-form-control"
                id="reg-password"
                name="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                autoComplete="new-password"
              />
            </div>
            <div className="mb-4">
              <label className="form-label mb-form-label" htmlFor="reg-password2">
                Подтверждение пароля
              </label>
              <input
                type="password"
                className="form-control mb-form-control"
                id="reg-password2"
                name="password_confirm"
                value={passwordConfirm}
                onChange={(e) => setPasswordConfirm(e.target.value)}
                autoComplete="new-password"
              />
            </div>
            <button type="submit" className="btn mb-btn-submit" disabled={submitting}>
              {submitting ? 'Регистрация…' : 'Зарегистрироваться'}
            </button>
          </form>
          <p className="text-center mt-3 mb-0 mb-link-muted">
            Уже есть аккаунт? <Link to="/bank/login">Войти</Link>
          </p>
        </div>
      </div>
    </>
  );
};

export default BankRegister;
