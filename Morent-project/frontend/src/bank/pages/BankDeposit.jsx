import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import toast from 'react-hot-toast';
import BankSimpleNav from '../components/BankSimpleNav';
import { BANK_UNAVAILABLE_MSG } from '../constants';

const BankDeposit = () => {
  const [amount, setAmount] = useState('');

  const handleSubmit = (e) => {
    e.preventDefault();
    if (!amount || Number(amount) <= 0) {
      toast.error('Укажите корректную сумму');
      return;
    }
    toast.error(BANK_UNAVAILABLE_MSG);
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
                Сумма, ₽
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
            <button type="submit" className="btn mb-btn-submit">
              Пополнить
            </button>
          </form>
        </div>
      </div>
    </>
  );
};

export default BankDeposit;
