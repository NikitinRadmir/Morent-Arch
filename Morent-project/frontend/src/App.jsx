import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import Home from './pages/Home';
import Category from './pages/Category';
import Layout from './layouts/Layout';
import CarDetail from './pages/CarDetail';
import PayPage from './pages/PayPage';
import SignIn from './pages/SignIn';
import SignUp from './pages/SignUp';
import VerifyEmail from './pages/VerifyEmail';
import Favorites from './pages/Favorites';
import Rentals from './pages/Rentals';
import Settings from './pages/Settings';
import Admin from './pages/Admin';
import RentalSuccess from './pages/RentalSuccess';
import NotFound from './pages/NotFound';
import BankLayout from './bank/BankLayout';
import BankAuthGate from './bank/components/BankAuthGate';
import BankHome from './bank/pages/BankHome';
import BankTransfer from './bank/pages/BankTransfer';
import BankDeposit from './bank/pages/BankDeposit';


const App = () => {
  return (
    <AuthProvider>
    <Router>
      <Routes>
        <Route path="/bank" element={<BankLayout />}>
          <Route element={<BankAuthGate />}>
            <Route index element={<BankHome />} />
            <Route path="transfer" element={<BankTransfer />} />
            <Route path="deposit" element={<BankDeposit />} />
          </Route>
          <Route path="login" element={<Navigate to="/bank" replace />} />
          <Route path="register" element={<Navigate to="/bank" replace />} />
        </Route>

        <Route path="/" element={<Layout />}>
          <Route index element={<Home />} />
          <Route path="category" element={<Category />} />
          <Route path="cardetail/:id" element={<CarDetail />} />
          <Route path="rent/:id" element={<PayPage />} />
          <Route path="sign-in" element={<SignIn />} />
          <Route path="sign-up" element={<SignUp />} />
          <Route path="verify-email" element={<VerifyEmail />} />
          <Route path="favorites" element={<Favorites />} />
          <Route path="rentals" element={<Rentals />} />
          <Route path="settings" element={<Settings />} />
          <Route path="rental-success" element={<RentalSuccess />} />
          <Route path="admin" element={<Admin />} />
          <Route path="*" element={<NotFound />} />
        </Route>
      </Routes>
    </Router>
    </AuthProvider>
  );
};

export default App;
