import React, { useState, useEffect } from 'react';
import PayForm from '../components/PayForm';
import RentalSum from '../components/RentalSum';
import { useParams } from 'react-router-dom';
import { API_BASE_URL } from '../context/AuthContext';



const PayPage = () => {
    const { id } = useParams();
    const [car, setCar] = useState(null); // Инициализируем как null
    const [totalAmount, setTotalAmount] = useState(0);
        
        // Загрузка данных конкретного автомобиля
        useEffect(() => {
            fetch(`${API_BASE_URL}/cars/${id}`)
                .then(async (response) => {
                    if (!response.ok) {
                        throw new Error(`HTTP error! status: ${response.status}`);
                    }
                    let data = await response.json();
                    setCar(data);
                    setTotalAmount(data.price);
                })
                .catch((error) => {
                    console.error("Error fetching car details:", error);
                });
        }, [id]);
    

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