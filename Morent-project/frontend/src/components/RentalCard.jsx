import React from 'react';
import { Link } from 'react-router-dom';
import CarImage from './CarImage';

const formatDate = (value) => {
    if (!value) return '-';
    const date = new Date(value);
    return date.toLocaleDateString();
};

const RentalCard = ({ rental }) => {
    const { car } = rental;
    return (
        <div className="rental-card">
            <div className="rental-card__image">
                <CarImage src={car.imgSrc} alt={car.name} fallbackKey={`${car.name} ${car.type}`} />
            </div>
            <div className="rental-card__content">
                <div className="rental-card__header">
                    <div>
                        <h4>{car.name}</h4>
                        <p className="highlited-gray mb-0">{car.type}</p>
                    </div>
                    <div className="rental-card__price">
                        ${rental.totalPrice.toFixed(2)}
                    </div>
                </div>
                <div className="rental-card__meta">
                    <div>
                        <span>Pick-up</span>
                        <strong>{formatDate(rental.startDate)}</strong>
                    </div>
                    <div>
                        <span>Drop-off</span>
                        <strong>{formatDate(rental.endDate)}</strong>
                    </div>
                </div>
                <div className="rental-card__actions">
                    <Link to={`/cardetail/${car.id}`} className="abutton">
                        View car
                    </Link>
                </div>
            </div>
        </div>
    );
};

export default RentalCard;

