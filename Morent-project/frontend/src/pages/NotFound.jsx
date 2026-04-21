import React from 'react';
import { Link } from 'react-router-dom';
import './NotFound.css';

const NotFound = () => (
  <div className="notfound-container">
    <h1 className="notfound-title">404</h1>
    <h2 className="notfound-subtitle">Страница не найдена</h2>
    <Link to="/" className="notfound-home-btn">На главную</Link>
  </div>
);

export default NotFound;









