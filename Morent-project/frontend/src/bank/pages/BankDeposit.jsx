import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import toast from 'react-hot-toast';
import BankSimpleNav from '../components/BankSimpleNav';
import { useBank } from '../context/BankContext';

const BankDeposit = () => {
  const navigate = useNavigate();
  const { deposit, isAuthenticated } = useBank();
  const [amount, setAmount] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!isAuthenticated) {
      toast.error('Банковский счёт недоступен');
      return;
    }
    if (!amount || Number(amount) <= 0) {
      toast.error('Укажите корректную сумму');
      return;
    }
    setSubmitting(true);
    try {
      await deposit(amount);
      navigate('/bank');
    } catch (err) {
      toast.error(err.message || 'Ошибка пополнения');
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
          <h1 className="mb-auth-title">Пополнение</h1>
          <form onSubmit={handleSubmit}>
            <div className="mb-4">
              <label className="form-label mb-form-label" htmlFor="deposit-amount">
                Сумма, $
              </label>
              <input
                type="number"
                className="form-control mb-form-control"
                id="deposit-amount"
                name="amount"
                min="1"
                step="0.01"
                placeholder="0"
                inputMode="decimal"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
              />
            </div>
            <button type="submit" className="btn mb-btn-submit" disabled={submitting}>
              {submitting ? 'Пополнение…' : 'Пополнить'}
            </button>
          </form>
        </div>
      </div>
    </>
  );
};

export default BankDeposit;
