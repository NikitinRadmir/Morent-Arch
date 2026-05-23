import React, { useState, useEffect } from 'react';
import CarsSection from '../components/CarsSection';
import Reviews from '../components/Reviews';
import CarInfo from '../components/CarInfo';
import { useParams } from 'react-router-dom';
import { graphqlRequest } from '../api/graphqlClient';

const CarDetail = () => {
    const [cars, setCars] = useState([]);
    const { id } = useParams();

    
    

    const [car, setCar] = useState(null); // Инициализируем как null
    const [reviews, setReview] = useState([]);

    // Загрузка данных через GraphQL (машина, комментарии, список машин)
    useEffect(() => {
        const load = async () => {
            try {
                const data = await graphqlRequest(
                    `
                    query CarDetail($id: ID!, $carId: Int!) {
                        car(id: $id) {
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
                        comments(carId: $carId) {
                            id
                            carId
                            userId
                            name
                            post
                            photo
                            date
                            rating
                            description
                        }
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
                    `,
                    { id, carId: Number(id) }
                );
                setCar(data.car || null);
                setReview(data.comments || []);
                setCars(data.cars || []);
            } catch (e) {
                console.error('Failed to load car detail via GraphQL', e);
            }
        };
        if (id) {
            load();
        }
    }, [id]);

    // Если автомобиль не найден
    if (!car) {
        return <div>Car not found</div>;
    }

    return (
        <div className="row">
            <div className="col-3 p-0 leftmenu-PC">
                {/* <LeftMenu AllCars={cars} /> */}
            </div>
            <div className="col-12 background-gray p-0 fake-col-12">
                <CarInfo props={car} />
                <Reviews
                    reviews={reviews}
                    carId={car.id}
                    onCommentAdded={(newComment) => setReview((prev) => [newComment, ...prev])}
                />
                <CarsSection cars={cars} col3 = {true}/>
            </div>
        </div>
    );
};

export default CarDetail;
