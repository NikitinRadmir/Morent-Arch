import React from 'react';
import { Outlet } from 'react-router-dom';
import { BankProvider } from './context/BankContext';
import '../styles/bank.css';

const BankLayout = () => {
  return (
    <BankProvider>
      <div className="bank-root mb-bg-pattern">
        <Outlet />
      </div>
    </BankProvider>
  );
};

export default BankLayout;
