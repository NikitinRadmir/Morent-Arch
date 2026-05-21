import React from 'react';
import { Outlet } from 'react-router-dom';
import '../styles/bank.css';

const BankLayout = () => {
  return (
    <div className="bank-root mb-bg-pattern">
      <Outlet />
    </div>
  );
};

export default BankLayout;
