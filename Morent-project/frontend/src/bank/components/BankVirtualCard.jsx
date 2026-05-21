import React from 'react';

const BankVirtualCard = ({ profile, loading }) => {
  if (loading) {
    return (
      <section className="mb-virtual-card mb-virtual-card--loading">
        <div className="mb-virtual-card-inner">
          <p className="mb-0">Загрузка карты…</p>
        </div>
      </section>
    );
  }

  if (!profile?.cardNumber) {
    return (
      <section className="mb-virtual-card">
        <div className="mb-virtual-card-inner mb-virtual-card-inner--empty">
          <p className="mb-0">Карта пока не выпущена. Обновите страницу или перезайдите в банк.</p>
        </div>
      </section>
    );
  }

  return (
    <section className="mb-virtual-card">
      <div className="mb-virtual-card-inner">
        <div className="mb-virtual-card-top">
          <span className="mb-virtual-card-chip" aria-hidden="true" />
          <span className="mb-virtual-card-brand">Morent</span>
        </div>
        <p className="mb-virtual-card-number">{profile.cardNumber}</p>
        <div className="mb-virtual-card-bottom">
          <div>
            <span className="mb-virtual-card-label">Card Holder</span>
            <span className="mb-virtual-card-value">{profile.cardHolder || '—'}</span>
          </div>
          <div>
            <span className="mb-virtual-card-label">Expires</span>
            <span className="mb-virtual-card-value">{profile.expDate || '—'}</span>
          </div>
          <div>
            <span className="mb-virtual-card-label">CVV</span>
            <span className="mb-virtual-card-value">{profile.cvv || '—'}</span>
          </div>
        </div>
      </div>
    </section>
  );
};

export default BankVirtualCard;
