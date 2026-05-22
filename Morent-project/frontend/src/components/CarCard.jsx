import React, { useContext } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { AuthContext } from '../context/AuthContext';
import CarImage from './CarImage';

const CarCard = ({ id, name, type, imgSrc, fuel, transmission, capacity, price }) => {
    const navigate = useNavigate();
    const { isAuthenticated, isFavorite, toggleFavorite } = useContext(AuthContext);
    const carId = Number(id);
    const isFav = isAuthenticated && Number.isFinite(carId) && isFavorite(carId);

    const handleFavoriteClick = async (e) => {
        e.preventDefault();
        e.stopPropagation();
        if (!isAuthenticated) {
            navigate('/sign-in');
            return;
        }
        try {
            await toggleFavorite(carId);
        } catch (error) {
            console.error('Failed to toggle favorite', error);
        }
    };

    return (
        <div className="card car-card p-3">
            <div className="row align-items-start">
                <div className="col-10">
                    <h5 className="name">{name}</h5>
                    <h6 className="car-class highlited-gray">{type}</h6>
                </div>
                <div className="col-2 d-flex justify-content-end">
                    <button type="button" className={`favorite-btn${isFav ? ' active' : ''}`} aria-label="add to favorite" onClick={handleFavoriteClick}>
                        <i className={isFav ? 'fa-solid fa-heart' : 'fa-regular fa-heart'}></i>
                    </button>
                </div>
            </div>
            <div className="car-img-container px-4">
                <CarImage src={imgSrc} className="car-img" alt={name} fallbackKey={`${name} ${type}`} />
            </div>
            <div className="row car-info mt-2">
                <div className="col-4 car-info-item">
                    <i className="fa-solid fa-gas-pump"></i>
                    <p className="highlited-gray mb-0">{fuel}L</p>
                </div>
                <div className="col-4 car-info-item">
                    <i className="fa-solid fa-gear"></i>
                    <p className="highlited-gray mb-0 text-capitalize">{transmission}</p>
                </div>
                <div className="col-4 car-info-item">
                    <i className="fa-solid fa-user"></i>
                    <p className="highlited-gray mb-0">{capacity} People</p>
                </div>
            </div>
            <div className="row mb-3 mt-4 align-items-center car-card__footer">
                <div className="col-7 p-0">
                    <p className="price text-start mb-0">${price}.00/<span className="highlited-gray">day</span></p>
                </div>
                <div className="col-5 text-end"><Link className='abutton' to={`/cardetail/${id}`}>Rent now</Link></div>
            </div>
        </div>
    );
};

export default CarCard;
