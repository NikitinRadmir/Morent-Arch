import React, { useState, useEffect } from 'react';
import PayForm from '../components/PayForm';
import RentalSum from '../components/RentalSum';
import { useParams } from 'react-router-dom';
import { API_BASE_URL } from '../context/AuthContext';
import { resilientJson } from '../utils/apiClient';



const PayPage = () => {
    const { id } = useParams();
    const [car, setCar] = useState(null); // Инициализируем как null
    const [totalAmount, setTotalAmount] = useState(0);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
        
        // Загрузка данных конкретного автомобиля
        useEffect(() => {
            let cancelled = false;
            setLoading(true);
            setError('');
            resilientJson(`${API_BASE_URL}/cars/${id}`)
                .then((data) => {
                    if (cancelled) return;
                    setCar(data);
                    setTotalAmount(data.price);
                })
                .catch((error) => {
                    if (cancelled) return;
                    console.error("Error fetching car details:", error);
                    setCar(null);
                    setError(error.message || 'Не удалось загрузить автомобиль');
                })
                .finally(() => {
                    if (!cancelled) setLoading(false);
                });
            return () => {
                cancelled = true;
            };
        }, [id]);
    
    if (loading) {
        return <div className="p-4">Loading...</div>;
    }

    if (error) {
        return <div className="alert alert-danger m-4">{error}</div>;
    }

    if (!car) {
        return <div>Car not found</div>;
    }

    return (
        <div className="row">
            <div className="col-12 order-2 order-xl-1 col-xl-7 p-0 background-gray">
                <PayForm car = {car} setTotalAmount={setTotalAmount}/>
            </div>
            <div className="col-12 order-1 order-xl-2 col-xl-5 p-0 background-gray">
                <RentalSum car = {car} totalAmount = {totalAmount}/>
            </div>
        </div>
    );
};

export default PayPage;
