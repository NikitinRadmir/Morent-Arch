import React from 'react';
import { Link, useLocation } from 'react-router-dom';

const RentalSuccess = () => {
    const location = useLocation();
    const car = location.state?.car;
    const rental = location.state?.rental;

    return (
        <div className="background-gray py-5">
            <div className="container">
                <div className="row justify-content-center">
                    <div className="col-12 col-md-8 col-lg-6">
                        <div className="text-center mb-4">
                            <div style={{ fontSize: '64px', color: '#3563e9', marginBottom: '20px' }}>✓</div>
                            <h2 className="mb-3">Аренда успешно оформлена!</h2>
                            <p className="highlited-gray mb-4">
                                Ваша аренда была успешно подтверждена. Детали отправлены на ваш email.
                            </p>
                        </div>

                        {car && (
                            <div className="form-container p-4 mb-4">
                                <h5 className="mb-3">Детали аренды</h5>
                                <div className="row">
                                    <div className="col-12 mb-3">
                                        <strong>Автомобиль:</strong> {car.name}
                                    </div>
                                    {rental && (
                                        <>
                                            <div className="col-12 mb-3">
                                                <strong>Дата начала:</strong> {new Date(rental.startDate).toLocaleDateString('ru-RU')}
                                            </div>
                                            <div className="col-12 mb-3">
                                                <strong>Дата окончания:</strong> {new Date(rental.endDate).toLocaleDateString('ru-RU')}
                                            </div>
                                            <div className="col-12 mb-3">
                                                <strong>Общая стоимость:</strong> ${rental.totalPrice?.toFixed(2) || '0.00'}
                                            </div>
                                        </>
                                    )}
                                </div>
                            </div>
                        )}

                        <div className="text-center">
                            <Link to="/rentals" className="auth-submit" style={{ display: 'inline-block', textDecoration: 'none', marginRight: '12px' }}>
                                Мои аренды
                            </Link>
                            <Link to="/" className="admin-btn admin-btn--ghost" style={{ display: 'inline-block', textDecoration: 'none' }}>
                                На главную
                            </Link>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default RentalSuccess;








