import React from 'react';
import Header from '../components/Header';
import Footer from '../components/Footer';
import { Outlet } from 'react-router-dom';
import { SearchProvider } from '../context/SearchContext';
import { MenuProvider } from '../context/MenuContext'; // Импортируйте MenuProvider
import { AuthProvider } from '../context/AuthContext';

const Layout = () => {
    return (
        <div>
            <AuthProvider>
                <SearchProvider>
                    <MenuProvider>
                        <Header />
                        <main>
                            <Outlet />
                        </main>
                        <Footer />
                    </MenuProvider>
                </SearchProvider>
            </AuthProvider>
        </div>
    );
};

export default Layout;
