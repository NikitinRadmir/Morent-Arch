import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import toast from 'react-hot-toast';
import BankSimpleNav from '../components/BankSimpleNav';
import PhoneInput from '../components/PhoneInput';
import { useBank } from '../context/BankContext';

const BankTransfer = () => {
  const navigate = useNavigate();
  const { transfer, isAuthenticated } = useBank();
  const [phone, setPhone] = useState('');
  const [amount, setAmount] = useState('');
  const [confirmed, setConfirmed] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!isAuthenticated) {
      toast.error('Войдите в банк');
      navigate('/bank/login');
      return;
    }
    if (!confirmed) {
      toast.error('Подтвердите перевод');
      return;
    }
    if (!phone.trim() || !amount || Number(amount) <= 0) {
      toast.error('Заполните телефон и сумму');
      return;
    }
    setSubmitting(true);
    try {
      await transfer(phone, amount);
      navigate('/bank');
    } catch (err) {
      toast.error(err.message || 'Ошибка перевода');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <>
      <BankSimpleNav
        extraLink={
          <Link to="/bank" className="mb-nav-muted-link">
            На главную
          </Link>
        }
      />
      <div className="mb-auth-wrapper">
        <div className="mb-auth-card">
          <h1 className="mb-auth-title">Перевод</h1>
          <form onSubmit={handleSubmit}>
            <div className="mb-3">
              <label className="form-label mb-form-label" htmlFor="transfer-phone">
                Телефон получателя
              </label>
              <PhoneInput
                id="transfer-phone"
                name="recipient_phone"
                value={phone}
                onChange={setPhone}
              />
            </div>
            <div className="mb-3">
              <label className="form-label mb-form-label" htmlFor="transfer-amount">
                Сумма к списанию, ₽
              </label>
              <input
                type="number"
                className="form-control mb-form-control"
                id="transfer-amount"
                name="amount"
                min="1"
                step="0.01"
                placeholder="0"
                inputMode="decimal"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
              />
            </div>
            <div className="form-check mb-4">
              <input
                className="form-check-input"
                type="checkbox"
                id="transfer-confirm"
                checked={confirmed}
                onChange={(e) => setConfirmed(e.target.checked)}
                required
              />
              <label className="form-check-label small text-secondary" htmlFor="transfer-confirm">
                Я уверен в своём решении перевести указанную сумму на указанный счёт
              </label>
            </div>
            <button type="submit" className="btn mb-btn-submit" disabled={submitting}>
              {submitting ? 'Перевод…' : 'Перевести'}
            </button>
          </form>
        </div>
      </div>
    </>
  );
};

export default BankTransfer;
