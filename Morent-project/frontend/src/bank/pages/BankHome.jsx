import React from 'react';
import { Link } from 'react-router-dom';
import BankHeader from '../components/BankHeader';
import { useBank } from '../context/BankContext';

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
  const { balance } = useBank();
  const balanceFormatted = balance.toLocaleString('ru-RU');

  return (
    <>
      <BankHeader />
      <main className="mb-container mb-5">
        <section className="mb-hero-balance text-center text-md-start">
          <p className="mb-hero-balance-label mb-0">Доступно на счёте</p>
          <p className="mb-hero-balance-value mb-0">
            {balanceFormatted}
            <span className="mb-hero-balance-currency">₽</span>
          </p>
        </section>

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

        <p className="text-center mb-link-muted mb-0">
          <Link to="/bank/login">Вход</Link>
          &nbsp;·&nbsp;
          <Link to="/bank/register">Регистрация</Link>
          &nbsp;·&nbsp;
          <Link to="/">На сайт MORENT</Link>
        </p>
      </main>
    </>
  );
};

export default BankHome;
