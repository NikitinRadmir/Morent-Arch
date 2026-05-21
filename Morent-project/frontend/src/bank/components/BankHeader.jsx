import React from 'react';
import { Link } from 'react-router-dom';
import { useBank } from '../context/BankContext';

const formatMoney = (value) => {
  if (value == null || Number.isNaN(Number(value))) {
    return '—';
  }
  return new Intl.NumberFormat('ru-RU', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(Number(value));
};

const BankHeader = () => {
  const { profile, isAuthenticated, logout, loading } = useBank();

  return (
    <header className="mb-header">
      <div className="mb-container">
        <div className="mb-row align-items-center">
          <div className="mb-col-md-3 mb-col-6">
            <Link className="mb-logo" to="/bank">
              <span className="mb-logo-mark">M</span>
              Morent Bank
            </Link>
          </div>
          <div className="mb-col-md-3 mb-col-6 text-end text-md-center">
            <div className="mb-header-balance">
              <span className="mb-header-balance-label">Баланс:</span>
              <strong className="mb-header-balance-value">
                {loading ? '…' : formatMoney(profile?.balance)}
              </strong>
            </div>
          </div>
          <div className="mb-col-12 mb-col-md-4 d-flex flex-wrap gap-2 mb-header-actions justify-content-md-center justify-content-start">
            <Link to="/bank/deposit" className="btn mb-btn-accent mb-btn-header">
              Пополнение
            </Link>
            <Link to="/bank/transfer" className="btn mb-btn-outline-light-custom mb-btn-header">
              Перевод
            </Link>
          </div>
          <div className="mb-col-12 mb-col-md-2 d-flex justify-content-md-end justify-content-end">
            <div className="mb-profile">
              <div>
                <div className="mb-profile-name">
                  {isAuthenticated ? profile?.displayName || profile?.phone : 'Гость'}
                </div>
                <div className="mb-profile-role">{profile?.role || 'Клиент'}</div>
              </div>
              {isAuthenticated ? (
                <button
                  type="button"
                  className="btn btn-link btn-sm text-white p-0 ms-2"
                  onClick={() => logout()}
                >
                  Выйти
                </button>
              ) : null}
              <img
                className="mb-profile-avatar"
                src="/images/bank/profile.png"
                width={44}
                height={44}
                alt="Фото профиля"
              />
            </div>
          </div>
        </div>
      </div>
    </header>
  );
};

export default BankHeader;
