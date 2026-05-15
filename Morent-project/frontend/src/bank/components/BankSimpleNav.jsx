import React from 'react';
import { Link } from 'react-router-dom';

const BankSimpleNav = ({ extraLink }) => {
  return (
    <nav className="mb-nav-simple">
      <div className="mb-container d-flex justify-content-between align-items-center">
        <Link className="mb-logo" to="/bank">
          <span className="mb-logo-mark">M</span>
          Morent Bank
        </Link>
        {extraLink}
      </div>
    </nav>
  );
};

export default BankSimpleNav;
