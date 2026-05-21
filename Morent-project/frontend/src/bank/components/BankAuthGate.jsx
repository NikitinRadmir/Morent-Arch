import React, { useContext } from 'react';
import { Navigate, Outlet } from 'react-router-dom';
import { AuthContext } from '../../context/AuthContext';

const BankAuthGate = () => {
  const { isAuthenticated } = useContext(AuthContext);

  if (!isAuthenticated) {
    return <Navigate to="/sign-in" replace state={{ from: '/bank' }} />;
  }

  return <Outlet />;
};

export default BankAuthGate;
