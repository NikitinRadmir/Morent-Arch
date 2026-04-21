import React, { useState } from 'react';
import CarCard from './CarCard';
import { Link } from 'react-router-dom';

const CarsSection = ({ cars, title, showViewAll = false, onViewAllClick, col3 = false, col4 = false }) => {

    const [visibleCars, setVisibleCars] = useState(12);
    const increment = 12;

    // Функция для показа всех отзывов
    const handleShowAll = () => {
        setVisibleCars(prevVisibleCars => prevVisibleCars + increment); 
    };


    return (
        <section className="cars-cards">
            {title && (
                <div className="row mb-3 align-items-center">
                    <h4 className="highlited-gray col-6">{title}</h4>
                    {showViewAll && (
                    <div className="col-6 d-flex justify-content-end">
                        <div
                            className="highlited-gray viewAlla"
                            onClick={onViewAllClick}
                            style={{ cursor: 'pointer' }}>
                            <Link to='/category'>View All</Link>
                        </div>
                    </div>
                    )}
                </div>
            )}


            <div className="row">
                {cars.length > 0 && col3 && (
                    cars.slice(0,visibleCars).map((car) => (
                        <div key={car.id} className="col-12 col-sm-6 col-lg-4 col-xl-3 car p-3">
                            <CarCard {...car} />
                        </div>
                    ))
                )} 
                {cars.length > 0 && col4 && (
                    cars.slice(0,visibleCars).map((car) => (
                        <div key={car.id} className="col-12 col-sm-6 col-lg-4 car p-3">
                            <CarCard {...car} />
                        </div>
                    ))
                )} 
                {cars.length <= 0 &&  (
                    <div className="col-12 d-flex justify-content-center py-5">
                        <div className="no-cars-card text-center">
                            <div className="no-cars-icon mb-3">
                                <i className="fa-regular fa-face-frown"></i>
                            </div>
                            <h4 className="mb-2">No cars found</h4>
                            <p className="highlited-gray mb-0">
                                Try changing your filters or search query.
                            </p>
                        </div>
                    </div>
                )} 
            </div>
            {title != "Popular cars" && visibleCars < cars.length ? (
                <div className='swap-button-div mt-4 pb-5'>
                    <button className='btn btn-primary' onClick={handleShowAll}>Show more cars</button>
                </div>
            ) : (
                <p></p>
            )}
            
        </section>
    );
};

export default CarsSection;