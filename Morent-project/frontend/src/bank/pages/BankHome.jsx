import React from 'react';
import { Link } from 'react-router-dom';
import BankHeader from '../components/BankHeader';
import BankVirtualCard from '../components/BankVirtualCard';
import { useBank } from '../context/BankContext';
import { formatUsd } from '../../utils/formatMoney';

const offers = [
  {
    badge: 'Кэшбэк',
    title: 'До 30% за покупки по карте',
    text: 'Подключите категории в приложении и получайте повышенный кэшбэк у партнёров Morent Bank.',
  },
  {
    badge: 'Вклад',
    title: 'Ставка до 16% годовых',
    text: 'Откройте вклад «Стабильный рост» онлайн — начисление процентов ежемесячно.',
  },
  {
    badge: 'Ипотека',
    title: 'Сниженный первый взнос',
    text: 'Программа для семей с детьми: одобрение за 48 часов и персональная ставка.',
  },
];

const BankHome = () => {
  const { profile, isAuthenticated, loading } = useBank();

  return (
    <>
      <BankHeader />
      <main className="mb-container mb-5">
        <section className="mb-hero-balance text-center text-md-start">
          <p className="mb-hero-balance-label mb-0">Доступно на счёте</p>
          <p className="mb-hero-balance-value mb-0">
            {loading ? '…' : formatUsd(profile?.balance)}
          </p>
        </section>

        {isAuthenticated && (
          <section className="mb-4">
            <h2 className="mb-section-title">Ваша карта</h2>
            <BankVirtualCard profile={profile} loading={loading} />
          </section>
        )}

        <section className="mb-4">
          <h2 className="mb-section-title">Акции и предложения</h2>
          <div className="mb-row g-3">
            {offers.map((offer) => (
              <div key={offer.title} className="mb-col-md-4 mb-col-12">
                <article className="mb-offer-card">
                  <span className="mb-offer-badge">{offer.badge}</span>
                  <h3 className="mb-offer-title">{offer.title}</h3>
                  <p className="mb-offer-text">{offer.text}</p>
                </article>
              </div>
            ))}
          </div>
        </section>

        {isAuthenticated && (
          <p className="text-center mb-link-muted mb-0">
            {profile?.displayName || profile?.phone}
            <br />
            <Link to="/">На сайт MORENT</Link>
          </p>
        )}
      </main>
    </>
  );
};

export default BankHome;
