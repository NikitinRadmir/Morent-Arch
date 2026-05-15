import React, { useContext, useState } from 'react';
import { SearchContext } from '../context/SearchContext';
import { Link, useNavigate } from 'react-router-dom';
import { useMenu } from '../context/MenuContext'; // Импортируйте useMenu
import { AuthContext } from '../context/AuthContext';
import { graphqlRequest } from '../api/graphqlClient';

const Header = () => {
    const { setSearchQuery } = useContext(SearchContext);
    const { toggleMenu } = useMenu(); // Используйте useMenu
    const { isAuthenticated, user, logout } = useContext(AuthContext);
    const [searchInput, setSearchInput] = useState('');
    const [suggestions, setSuggestions] = useState([]);
    const [isProfileMenuOpen, setIsProfileMenuOpen] = useState(false);
    const navigate = useNavigate();

    const handleSearchChange = async (e) => {
        const newValue = e.target.value;
        setSearchInput(newValue);

        if (newValue.trim()) {
            try {
                const data = await graphqlRequest(
                    `
                    query SearchCars($name: String!) {
                        cars {
                            id
                            name
                        }
                    }
                    `
                );
                const all = data.cars || [];
                const lowered = newValue.toLowerCase();
                const filtered = all.filter((car) =>
                    car.name.toLowerCase().includes(lowered)
                );
                setSuggestions(filtered.slice(0, 5));
            } catch (error) {
                console.error('Error fetching cars via GraphQL:', error);
                setSuggestions([]);
            }
        } else {
            setSuggestions([]);
        }

        setSearchQuery(newValue);
    };

    // Навигация по Enter
    const handleKeyDown = (e) => {
        if (e.key === 'Enter') {
            if(searchInput.trim()) {
                navigate(`/category?q=${encodeURIComponent(searchInput.trim())}`);
                setSuggestions([]);
            }
        }
    };

    // Навигация по клику на лупу
    const handleSearchClick = () => {
        if(searchInput.trim()) {
            navigate(`/category?q=${encodeURIComponent(searchInput.trim())}`);
            setSuggestions([]);
        }
    };

    const handleFavoritesClick = () => {
        if (!isAuthenticated) {
            navigate('/sign-in');
            return;
        }
        navigate('/favorites');
    };

    const handleRentalsClick = () => {
        if (!isAuthenticated) {
            navigate('/sign-in');
            return;
        }
        navigate('/rentals');
    };

    const handleProfileClick = () => {
        if (!isAuthenticated) {
            navigate('/sign-in');
            return;
        }
        setIsProfileMenuOpen((prev) => !prev);
    };

    const avatarUrl = user?.avatarUrl || 'https://i.pravatar.cc/100?img=12';
    const displayName = user?.nickname || user?.name || user?.email;

    return (
        <header>
            <div className='site-header row headerPC'>
                <div className="col-2 d-flex justify-content-center">
                    <Link to="/"><h1 className="title">MORENT</h1></Link>
                </div>
                <div className='col-5 d-flex align-items-center search-wrapper'>
                    <div className="search-container">
                        <input
                            type="text"
                            className="search-input"
                            placeholder="Search something here"
                            value={searchInput}
                            onChange={handleSearchChange}
                            onKeyDown={handleKeyDown}
                        />
                        {suggestions.length > 0 && (
                            <div className="suggestions-container">
                                {suggestions.map((car) => (
                                    <Link
                                        key={car.id}
                                        to={`/CarDetail/${car.id}`}
                                        className="suggestion-item"
                                        onClick={() => {
                                            setSearchInput('');
                                            setSuggestions([]);
                                        }}
                                    >
                                        {car.name}
                                    </Link>
                                ))}
                            </div>
                        )}
                    </div>
                    <button type="button" className="search-btn ml-3" onClick={handleSearchClick} aria-label="Поиск">
                        <i className="fa-solid fa-magnifying-glass"></i>
                    </button>
                </div>
                <div className='col-3 offset-2 d-flex justify-content-center'>
                    <div className="Profile_And_Notification">
                        {!isAuthenticated ? (
                            <Link className="sign-in-btn" to="/sign-in">
                                Sign in
                            </Link>
                        ) : (
                            <div className="header-profile" onClick={handleProfileClick}>
                                <div className="profile">
                                    <img src={avatarUrl} alt="Profile" className="profile-image" />
                                </div>
                                <span className="header-profile-name">{displayName}</span>
                                <i className={`fa-solid fa-chevron-${isProfileMenuOpen ? 'up' : 'down'}`}></i>
                                {isProfileMenuOpen && (
                                    <div className="header-profile-menu">
                                        <button
                                            type="button"
                                            className="header-profile-menu-item"
                                            onClick={() => {
                                                setIsProfileMenuOpen(false);
                                                navigate('/settings');
                                            }}
                                        >
                                            Profile settings
                                        </button>
                                        {(() => {
                                            const userRole = user?.role || user?.Role;
                                            return userRole === 'admin';
                                        })() && (
                                            <button
                                                type="button"
                                                className="header-profile-menu-item"
                                                onClick={() => {
                                                    setIsProfileMenuOpen(false);
                                                    navigate('/admin');
                                                }}
                                            >
                                                Admin panel
                                            </button>
                                        )}
                                        <button
                                            type="button"
                                            className="header-profile-menu-item header-profile-menu-item--danger"
                                            onClick={() => {
                                                setIsProfileMenuOpen(false);
                                                logout();
                                            }}
                                        >
                                            Logout
                                        </button>
                                    </div>
                                )}
                            </div>
                        )}
                        <button className="like" type="button" onClick={handleFavoritesClick}>
                            <img src="/images/heart.png" alt="Like" />
                        </button>
                        <button className="rentals-btn" type="button" onClick={handleRentalsClick}>
                            <i className="fa-solid fa-car-side"></i>
                        </button>
                        <Link className="sign-in-btn" to="/bank" title="Morent Bank">
                            Bank
                        </Link>
                    </div>
                </div>
            </div>

            <div className='site-header row headerMobile'>
                <div className="col-7 pl-5">
                    <Link to="/"><h1 className="title">MORENT</h1></Link>
                </div>
                <div className='col-5 d-flex justify-content-end pr-5'>
                    <div className="Profile_And_Notification2">
                        {!isAuthenticated ? (
                            <Link className="sign-in-btn sign-in-btn--small" to="/sign-in">
                                Sign in
                            </Link>
                        ) : (
                            <button type="button" className="profile profile-button" onClick={handleProfileClick} aria-label="Profile menu">
                                <img src={avatarUrl} alt="Profile" className="profile-image" />
                            </button>
                        )}
                        {isProfileMenuOpen && isAuthenticated && (
                            <div className="header-profile-menu header-profile-menu--mobile">
                                <div className="header-profile-menu-header">
                                    <img src={avatarUrl} alt="Profile" className="profile-image" />
                                    <div>
                                        <div className="header-profile-name">{displayName}</div>
                                        {user?.position && <div className="header-profile-position">{user.position}</div>}
                                    </div>
                                </div>
                                <button
                                    type="button"
                                    className="header-profile-menu-item"
                                    onClick={() => {
                                        setIsProfileMenuOpen(false);
                                        navigate('/settings');
                                    }}
                                >
                                    Profile settings
                                </button>
                                {(() => {
                                    const userRole = user?.role || user?.Role;
                                    return userRole === 'admin';
                                })() && (
                                    <button
                                        type="button"
                                        className="header-profile-menu-item"
                                        onClick={() => {
                                            setIsProfileMenuOpen(false);
                                            navigate('/admin');
                                        }}
                                    >
                                        Admin panel
                                </button>
                                )}
                                <button
                                    type="button"
                                    className="header-profile-menu-item header-profile-menu-item--danger"
                                    onClick={() => {
                                        setIsProfileMenuOpen(false);
                                        logout();
                                    }}
                                >
                                    Logout
                                </button>
                            </div>
                        )}
                        <button className="rentals-btn rentals-btn--small" type="button" onClick={handleRentalsClick}>
                            <i className="fa-solid fa-car-side"></i>
                        </button>
                    </div>
                </div>
                <div className='col-12 px-5'>
                    <div className='row'>
                        <input
                            type="text"
                            className="search-input2"
                            placeholder="Search something here"
                            value={searchInput}
                            onChange={handleSearchChange}
                            onKeyDown={handleKeyDown}
                        />
                        <button type="button" className='search-btn' onClick={handleSearchClick} aria-label="Поиск">
                            <i className="fa-solid fa-magnifying-glass"></i>
                        </button>
                        {suggestions.length > 0 && (
                            <div className="suggestions-container">
                                {suggestions.map((car) => (
                                    <Link
                                        key={car.id}
                                        to={`/CarDetail/${car.id}`}
                                        className="suggestion-item"
                                        onClick={() => {
                                            setSearchInput('');
                                            setSuggestions([]);
                                        }}
                                    >
                                        {car.name}
                                    </Link>
                                ))}
                            </div>
                        )}
                    </div>
                </div>
            </div>

            <div className='site-header row headerMobileDetail'>
                <div className='col-7 pl-5'>
                    <button className='leftmenu-button' onClick={toggleMenu}><i className='fas fa-bars'></i></button> {/* Добавьте обработчик */}
                </div>
                <div className='col-5 d-flex justify-content-end pr-5'>
                    <div className="Profile_And_Notification2">
                        {!isAuthenticated ? (
                            <Link className="sign-in-btn sign-in-btn--small" to="/sign-in">
                                Sign in
                            </Link>
                        ) : (
                            <button type="button" className="profile profile-button" onClick={handleProfileClick} aria-label="Profile menu">
                                <img src={avatarUrl} alt="Profile" className="profile-image" />
                            </button>
                        )}
                        {isProfileMenuOpen && isAuthenticated && (
                            <div className="header-profile-menu header-profile-menu--mobile">
                                <div className="header-profile-menu-header">
                                    <img src={avatarUrl} alt="Profile" className="profile-image" />
                                    <div>
                                        <div className="header-profile-name">{displayName}</div>
                                        {user?.position && <div className="header-profile-position">{user.position}</div>}
                                    </div>
                                </div>
                                <button
                                    type="button"
                                    className="header-profile-menu-item"
                                    onClick={() => {
                                        setIsProfileMenuOpen(false);
                                        navigate('/settings');
                                    }}
                                >
                                    Profile settings
                                </button>
                                {(() => {
                                    const userRole = user?.role || user?.Role;
                                    return userRole === 'admin';
                                })() && (
                                    <button
                                        type="button"
                                        className="header-profile-menu-item"
                                        onClick={() => {
                                            setIsProfileMenuOpen(false);
                                            navigate('/admin');
                                        }}
                                    >
                                        Admin panel
                                </button>
                                )}
                                <button
                                    type="button"
                                    className="header-profile-menu-item header-profile-menu-item--danger"
                                    onClick={() => {
                                        setIsProfileMenuOpen(false);
                                        logout();
                                    }}
                                >
                                    Logout
                                </button>
                            </div>
                        )}
                        <button className="rentals-btn rentals-btn--small" type="button" onClick={handleRentalsClick}>
                            <i className="fa-solid fa-car-side"></i>
                        </button>
                    </div>
                </div>
                <div className="col-12 pl-5">
                    <Link to="/"><h1 className="title">MORENT</h1></Link>
                </div>
                <div className='col-12 px-5'>
                    <div className='row '>
                        <input
                            type="text"
                            className="search-input2"
                            placeholder="Search something here"
                            onChange={handleSearchChange}
                            onKeyDown={handleKeyDown}
                        />
                        <button type="button" className='search-btn' onClick={handleSearchClick} aria-label="Поиск">
                            <i className="fa-solid fa-magnifying-glass"></i>
                        </button>
                    </div>
                </div>
            </div>
        </header>
    );
};

export default Header;
