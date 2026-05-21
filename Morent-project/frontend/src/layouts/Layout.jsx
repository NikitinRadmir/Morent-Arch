import React from 'react';
import Header from '../components/Header';
import Footer from '../components/Footer';
import { Outlet } from 'react-router-dom';
import { SearchProvider } from '../context/SearchContext';
import { MenuProvider } from '../context/MenuContext'; // Импортируйте MenuProvider

const Layout = () => {
    return (
        <div>
            <SearchProvider>
                <MenuProvider>
                    <Header />
                    <main>
                        <Outlet />
                    </main>
                    <Footer />
                </MenuProvider>
            </SearchProvider>
        </div>
    );
};

export default Layout;
