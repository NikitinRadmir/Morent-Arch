import React, { useContext, useState, useEffect } from 'react';
import { SearchContext } from '../context/SearchContext';
import { useSearchParams } from 'react-router-dom';
import LeftMenu from '../components/LeftMenu';
import Options from '../components/Options';
import CarsSection from '../components/CarsSection';
import { useMenu } from '../context/MenuContext'; // Импортируйте useMenu
import { graphqlRequest } from '../api/graphqlClient';

const Category = () => {
    const { searchQuery, setSearchQuery } = useContext(SearchContext);
    const { isMenuOpen, toggleMenu } = useMenu(); 
    const [allCars, setAllCars] = useState([]);
    const [filteredCars, setFilteredCars] = useState([]);
    const [searchParams, setSearchParams] = useSearchParams();

    // Инициализация фильтров из URL
    const [filters, setFilters] = useState({
        types: searchParams.getAll('carType') || [],
        capacities: searchParams.getAll('capacity') || [],
        maxPrice: searchParams.get('priceUnder') || '',
    });

    // Синхронизация searchQuery с URL
    useEffect(() => {
        const queryFromUrl = searchParams.get('q') || '';
        if (queryFromUrl) {
            setSearchQuery(queryFromUrl);
        }
    }, [searchParams, setSearchQuery]);

    // Обновление URL при изменении searchQuery
    useEffect(() => {
        if (searchQuery) {
            searchParams.set('q', searchQuery);
        } else {
            searchParams.delete('q');
        }
        setSearchParams(searchParams);
    }, [searchQuery, searchParams, setSearchParams]);

    // Обновление URL при изменении фильтров
    useEffect(() => {
        searchParams.delete('carType');
        searchParams.delete('capacity');
        searchParams.delete('priceUnder');

        filters.types.forEach(type => searchParams.append('carType', type));
        filters.capacities.forEach(capacity => searchParams.append('capacity', capacity));
        if (filters.maxPrice) {
            searchParams.set('priceUnder', filters.maxPrice);
        }

        setSearchParams(searchParams);
    }, [filters, searchParams, setSearchParams]);

    // Загрузка всех машин (GraphQL)
    useEffect(() => {
        const load = async () => {
            try {
                const data = await graphqlRequest(`
                    query {
                        cars {
                            id
                            name
                            type
                            capacity
                            price
                            fuel
                            transmission
                            imgSrc
                            description
                        }
                    }
                `);
                const cars = data.cars || [];
                setAllCars(cars);
                // Применяем фильтры и поиск к данным только после загрузки
                applyFiltersAndSearch(cars);
            } catch (e) {
                console.error('Failed to load cars via GraphQL (category)', e);
            }
        };
        load();
    }, []);

    // Применение фильтров и поиска (логическое И)
    const applyFiltersAndSearch = (data) => {
        let filteredData = data;

        // Применяем фильтры
        filteredData = filteredData.filter(car => {
            const matchesType = filters.types.length === 0 || filters.types.includes(car.type);
            const matchesCapacity = filters.capacities.length === 0 || filters.capacities.includes(String(car.capacity));
            const matchesPrice = !filters.maxPrice || car.price <= Number(filters.maxPrice);

            return matchesType && matchesCapacity && matchesPrice;
        });

        // Применяем поиск
        if (searchQuery.trim() !== "") {
            filteredData = filteredData.filter(car =>
                car.name.toLowerCase().includes(searchQuery.toLowerCase())
            );
        }

        setFilteredCars(filteredData);
    };

    // Реакция на изменения фильтров и поисковика
    useEffect(() => {
        if (allCars.length > 0) {
            applyFiltersAndSearch(allCars);
        }
    }, [filters, searchQuery, allCars]);

    // Обработка изменения фильтров
    const handleFilterChange = (newFilters) => {
        setFilters(newFilters);

        const queryParams = new URLSearchParams();
        if (newFilters.types.length > 0) {
            newFilters.types.forEach(type => queryParams.append('carType', type));
        }
        if (newFilters.capacities.length > 0) {
            newFilters.capacities.forEach(capacity => queryParams.append('capacity', capacity));
        }
        queryParams.append('priceUnder', newFilters.maxPrice);

        // Пока фильтрация реализована на стороне клиента,
        // поэтому просто применяем applyFiltersAndSearch к allCars.
        if (allCars.length > 0) {
            applyFiltersAndSearch(allCars);
        }
    };

    return (
        <div className="row">
            <div className="col-3 p-0 leftmenu-PC">
                <LeftMenu props={allCars} onFilterChange={handleFilterChange} initialFilters={filters} />
            </div>
            {isMenuOpen && (
                <div className="col-8 p-0 leftmenu-mobile pt-5">
                    <LeftMenu props={allCars} onFilterChange={handleFilterChange} initialFilters={filters} />
                    <button className='btn btn-secondary ml-3' onClick={toggleMenu}>Close</button>
                </div>
                
            )}
            <div className="col-9 p-0 background-gray fake-col-12">
                <Options />
                <CarsSection cars={filteredCars} col4={true} />
            </div>
            {isMenuOpen && <div className="leftmenu-mobile-overlay" onClick={toggleMenu}></div>}
        </div>
    );
};

export default Category;
