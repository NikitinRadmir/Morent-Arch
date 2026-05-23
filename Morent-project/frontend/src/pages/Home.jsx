import React, { useState, useEffect } from 'react';
import Ads from '../components/Ads';
import Options from '../components/Options';
import CarsSection from '../components/CarsSection';
import { graphqlRequest } from '../api/graphqlClient';

const Home = () => {
    const [cars, setCars] = useState([]);
    const [filteredPopularCars, setFilteredPopularCars] = useState([]);
    const [filteredRecommendationCars, setFilteredRecommendationCars] = useState([]);

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
                setCars(data.cars || []);
            } catch (e) {
                console.error('Failed to load cars via GraphQL', e);
            }
        };
        load();
    }, []);

    useEffect(() => {
        if (cars.length > 0) {
            applySearch(cars);
        }
    }, [cars]);

    const applySearch = (data) => {
        setFilteredPopularCars(data.slice(0, 4));
        setFilteredRecommendationCars(data);
    };

    return (
        <div className='background-gray'>
            <Ads />
            <Options />
            <CarsSection
                cars={filteredPopularCars}
                title="Popular cars"
                showViewAll={true}
                onViewAllClick={() => console.log('View all popular cars')}
                col3 = {true}
            />
            <CarsSection
                cars={filteredRecommendationCars}
                title="Recommendation cars"
                showViewAll={false}
                col3 = {true}
            />
        </div>
    );
};

export default Home;
