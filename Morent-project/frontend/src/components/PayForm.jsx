import React, { useContext, useEffect, useMemo, useState } from 'react';
import toast from 'react-hot-toast';
import { useNavigate } from 'react-router-dom';
import { AuthContext, API_BASE_URL } from '../context/AuthContext';

const PayForm = ({ car, setTotalAmount }) => {
    const { isAuthenticated, authRequest } = useContext(AuthContext);
    const navigate = useNavigate();
    const [localTotalAmount, setLocalTotalAmount] = useState(car.price);
    const [submitting, setSubmitting] = useState(false);
    const [submitMessage, setSubmitMessage] = useState('');
    const [submitError, setSubmitError] = useState('');
    const [notifications, setNotifications] = useState([]);
    const [bookings, setBookings] = useState([]);
    // Массив всех занятых дат в виде строк YYYY-MM-DD (пересчитывается при bookings)
    const [bookedDateStrings, setBookedDateStrings] = useState([]);

    const today = useMemo(() => {
        const base = new Date();
        base.setHours(0, 0, 0, 0);
        return base;
    }, []);
    const [currentMonth, setCurrentMonth] = useState(new Date(today.getFullYear(), today.getMonth(), 1));
    const [selectedStartDate, setSelectedStartDate] = useState(null);
    const [selectedEndDate, setSelectedEndDate] = useState(null);
    const weekDays = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];

    const formatDate = (date) => (date ? date.toISOString().split('T')[0] : '');

    // Добавить функцию нормализации дат:
    const normalizeDateToISO = (date) => {
        if (!date) return '';
        const d = new Date(date);
        d.setHours(12, 0, 0, 0);
        return d.toISOString().split('T')[0];
    };

    // True если данная дата среди занятых
    const isDateInBookedRange = (date) => {
        if (!bookedDateStrings || bookedDateStrings.length === 0) return false;
        return bookedDateStrings.includes(normalizeDateToISO(date));
    };

    const calendarDays = useMemo(() => {
        const result = [];
        const year = currentMonth.getFullYear();
        const month = currentMonth.getMonth();
        const firstDay = new Date(year, month, 1);
        const startDay = firstDay.getDay();
        for (let i = 0; i < startDay; i++) {
            result.push(null);
        }
        const daysInMonth = new Date(year, month + 1, 0).getDate();
        for (let day = 1; day <= daysInMonth; day++) {
            const date = new Date(year, month, day);
            const disabledBase = date < today;
            const disabledByBooking = isDateInBookedRange(date);
            result.push({ day, date, disabled: disabledBase || disabledByBooking, booked: disabledByBooking });
        }
        while (result.length % 7 !== 0) {
            result.push(null);
        }
        return result;
    }, [currentMonth, today, bookedDateStrings]);

    const handlePrevMonth = () => {
        if (currentMonth.getFullYear() === today.getFullYear() && currentMonth.getMonth() === today.getMonth()) {
            return;
        }
        setCurrentMonth(new Date(currentMonth.getFullYear(), currentMonth.getMonth() - 1, 1));
    };

    const handleNextMonth = () => {
        setCurrentMonth(new Date(currentMonth.getFullYear(), currentMonth.getMonth() + 1, 1));
    };

    const resetTotalsToBase = () => {
        setLocalTotalAmount(car.price);
        setTotalAmount(car.price);
    };

    // Проверка, пересекается ли диапазон дат с занятыми датами
    const hasBookingConflict = (startDate, endDate) => {
        if (!startDate || !endDate || bookedDateStrings.length === 0) return false;
        // Проверяем каждую дату в диапазоне
        const checkDate = new Date(startDate);
        checkDate.setHours(12, 0, 0, 0);
        const endCheck = new Date(endDate);
        endCheck.setHours(12, 0, 0, 0);
        while (checkDate <= endCheck) {
            if (bookedDateStrings.includes(normalizeDateToISO(checkDate))) {
                return true;
            }
            checkDate.setDate(checkDate.getDate() + 1);
        }
        return false;
    };

    const handleDayClick = (item) => {
        if (!item || item.disabled) return;
        const date = item.date;

        if (!selectedStartDate || (selectedStartDate && selectedEndDate)) {
            setSelectedStartDate(date);
            setSelectedEndDate(null);
            resetTotalsToBase();
            return;
        }

        if (date < selectedStartDate) {
            setSelectedStartDate(date);
            setSelectedEndDate(null);
            resetTotalsToBase();
            return;
        }

        // Проверяем, не пересекается ли выбранный диапазон с занятыми датами
        if (hasBookingConflict(selectedStartDate, date)) {
            addNotification('Выбранный диапазон содержит занятые даты. Пожалуйста, выберите другой период.');
            return;
        }

        setSelectedEndDate(date);
    };

    const isSelectedDate = (item) => {
        if (!item || !item.date) return false;
        const formatted = formatDate(item.date);
        return (selectedStartDate && formatDate(selectedStartDate) === formatted) ||
            (selectedEndDate && formatDate(selectedEndDate) === formatted);
    };

    const isInRange = (item) => {
        if (!item || !item.date || !selectedStartDate || !selectedEndDate) return false;
        return item.date > selectedStartDate && item.date < selectedEndDate;
    };

    const addNotification = (message) => {
        const newNotification = { id: Date.now(), message };
        setNotifications((prev) => [...prev, newNotification]);

        setTimeout(() => {
            removeNotification(newNotification.id);
        }, 50000);
    };

    const removeNotificationByText = (message) => {
        setNotifications((prev) => prev.filter((n) => n.message !== message));
    };

    const removeNotification = (idOrMessage) => {
        if (typeof idOrMessage === 'string') {
            removeNotificationByText(idOrMessage);
        } else {
            setNotifications((prev) => prev.filter((n) => n.id !== idOrMessage));
        }
    };
    const calculateTotalAmount = (datePick, dateDrop) => {
        if (!datePick || !dateDrop) return;

        const pickDate = new Date(datePick);
        const dropDate = new Date(dateDrop);

        const diffInMs = dropDate - pickDate;
        const diffInDays = diffInMs / (1000 * 60 * 60 * 24);

        if (diffInDays <= 0) {
            setLocalTotalAmount(car.price);
            setTotalAmount(car.price);
            return;
        }

        const total = diffInDays * car.price;
        setLocalTotalAmount(total);
        setTotalAmount(total);
    };

    useEffect(() => {
        if (selectedStartDate && selectedEndDate) {
            calculateTotalAmount(formatDate(selectedStartDate), formatDate(selectedEndDate));
        }
    }, [selectedStartDate, selectedEndDate]);

    // загрузка бронирований для машины
    useEffect(() => {
        const loadBookings = async () => {
            try {
                const resp = await fetch(`${API_BASE_URL}/rentals/car/${car.id}`);
                if (!resp.ok) {
                    return;
                }
                const data = await resp.json();
                setBookings(data || []);
            } catch (e) {
                console.error('Failed to load bookings', e);
            }
        };
        if (car?.id) {
            loadBookings();
        }
    }, [car]);

    useEffect(() => {
        // Для каждого бронирования включаем в массив все даты в диапазоне [startDate, endDate] включительно
        const result = [];
        bookings.forEach(b => {
            const d0 = new Date(b.startDate);
            const d1 = new Date(b.endDate);
            d0.setHours(12,0,0,0);
            d1.setHours(12,0,0,0);
            let dt = new Date(d0);
            while (dt <= d1) {
                result.push(normalizeDateToISO(dt));
                dt.setDate(dt.getDate() + 1);
            }
        });
        setBookedDateStrings(result);
    }, [bookings]);


    const handleNameChange = (e) => {
        const value = e.target.value;
        const filteredValue = value.replace(/[^a-zA-Zа-яА-Я\s]/g, '');
        e.target.value = filteredValue;
    };

    const handleCardChange = (e) => {
        const value = e.target.value;
        const filteredValue = value.replace(/[^1-9\s]/g, '');
        e.target.value = filteredValue;
    };
    const handleCVCChange = (e) => {
        let value = e.target.value;
        let filteredValue = value.replace(/[^1-9\s]/g, '');
        if (filteredValue.length > 3) {
            filteredValue = filteredValue.slice(0, 3);
        }
        e.target.value = filteredValue;
    };

    const validateEmail = (e) => {
        const email = e.target.value;
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (!emailRegex.test(email)) {
            e.target.setCustomValidity('Please enter a valid email address.');
        } else {
            e.target.setCustomValidity('');
        }
    };


    // локальное состояние валидации срока карты пока не используется UI
    // eslint-disable-next-line no-unused-vars
    const [expirationDateError, setExpirationDateError] = useState('');

    const validateExpirationDate = (value) => {
        let day = value.slice(0, 2);
        let month = value.slice(3, 5);
        let year = value.slice(6, 8);
        removeNotification('Invalid day. Day must be between 01 and 31.');
        removeNotification('Invalid month. Month must be between 01 and 12.');
        removeNotification('Invalid year. Year must be between 00 and 99.');
        if (day && (parseInt(day, 10) < 1 || parseInt(day, 10) > 31)) {
            addNotification('Invalid day. Day must be between 01 and 31.');
            return false;
        }

        if (month && (parseInt(month, 10) < 1 || parseInt(month, 10) > 12)) {
            addNotification('Invalid month. Month must be between 01 and 12.');
            return false;
        }

        if (year && (parseInt(year, 10) < 0 || parseInt(year, 10) > 99)) {
            addNotification('Invalid year. Year must be between 00 and 99.');
            return false;
        }

        return true;
    };

    const handleExpirationDateChange = (e) => {
        let value = e.target.value;

        value = value.replace(/[^0-9]/g, '');

        if (value.length > 6) {
            value = value.slice(0, 6);
        }

        const day = value.slice(0, 2);
        const month = value.slice(2, 4);
        const year = value.slice(4, 6);

        let formattedValue = '';
        if (day) {
            formattedValue += day;
            if (value.length >= 3) {
                formattedValue += '/';
            }
        }
        if (month) {
            formattedValue += month;
            if (value.length >= 5) {
                formattedValue += '/';
            }
        }
        if (year) {
            formattedValue += year;
        }

        validateExpirationDate(formattedValue);

        e.target.value = formattedValue;
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        setSubmitMessage('');
        setSubmitError('');

        if (!isAuthenticated) {
            addNotification('Please sign in to create a rental.');
            toast.error('Пожалуйста, войдите в аккаунт, чтобы оформить аренду');
            return;
        }

        if (!selectedStartDate || !selectedEndDate) {
            addNotification('Please select pick-up and drop-off dates.');
            toast.error('Выберите даты начала и окончания аренды');
            return;
        }

        // check overlap with existing bookings
        const selStart = new Date(formatDate(selectedStartDate));
        const selEnd = new Date(formatDate(selectedEndDate));
        const overlaps = bookings.some((b) => {
            const bStart = new Date(b.startDate);
            const bEnd = new Date(b.endDate);
            bStart.setHours(0, 0, 0, 0);
            bEnd.setHours(0, 0, 0, 0);
            return selStart < bEnd && selEnd > bStart;
        });
        if (overlaps) {
            addNotification('Selected dates overlap with an existing booking. Please choose different days.');
            toast.error('Выбранные даты пересекаются с уже забронированными');
            return;
        }

        try {
            setSubmitting(true);
            const rentalPayload = {
                carId: car.id,
                startDate: normalizeDateToISO(selectedStartDate),
                endDate: normalizeDateToISO(selectedEndDate),
                totalPrice: localTotalAmount,
            };
            const rentalResponse = await authRequest('/rentals', {
                method: 'POST',
                body: JSON.stringify(rentalPayload),
            });

            // Optional: keep existing email notification flow
            const formData = new FormData(e.target);
            fetch(`${API_BASE_URL}/SendEmail/SendEmail/${car.id}`, {
                method: 'POST',
                body: formData,
            }).catch(() => {});

            // Redirect to success page
            toast.success('Бронирование успешно оформлено');
            navigate('/rental-success', {
                state: {
                    car,
                    rental: rentalResponse,
                },
            });
        } catch (error) {
            const msg = error.message || 'Failed to submit rental';
            setSubmitError(msg);
            addNotification(msg);
            toast.error(`Ошибка бронирования: ${msg}`);
        } finally {
            setSubmitting(false);
        }
    };

    return (
        <div>

            <div className="notification-container">
                {notifications.map((n, index) => (
                    <div
                        key={n.id}
                        className="floating-notification"
                        style={{ top: `${20 + index * 80}px` }}
                    >
                        <p>{n.message}</p>
                    </div>
                ))}
            </div>

            {submitMessage && <div className="alert alert-success mx-4">{submitMessage}</div>}
            {submitError && <div className="alert alert-danger mx-4">{submitError}</div>}

            <form onSubmit={handleSubmit} encType="multipart/form-data">
                <div className='row p-4'>
                    <div className='col-12 p-4 form-container'>
                        <div className='row p-0'>
                            <div className='col-12'>
                                <h5>Billing Info</h5>
                            </div>
                            <div className='col-8 pr-4'>
                                <p><highlited-gray>Please enter your billing info</highlited-gray></p>
                            </div>
                            <div className='col-4 pb-4 pl-4 t-e'>
                                <p><highlited-gray>Step 1 of 4</highlited-gray></p>
                            </div>
                            <div className='col-6 pr-4'>
                                <h6>Name</h6>
                                <input
                                    className="form-container-input"
                                    type="text"
                                    id="Name"
                                    name="Name"
                                    required placeholder="Your Name"
                                    onInput={handleNameChange}
                                />
                            </div>
                            <div className='col-6 pb-4 pl-4'>
                                <h6>Email</h6>
                                <input
                                    className="form-container-input"
                                    type="text"
                                    id="Email"
                                    name="Email"
                                    required placeholder="Email"
                                    onInput={validateEmail}
                                />
                            </div>
                            <div className='col-6 pr-4'>
                                <h6>Address</h6>
                                <input className="form-container-input" type="text" id="Address" name="Address" required placeholder="Address"></input>
                            </div>
                            <div className='col-6 pl-4'>
                                <h6>Town/City</h6>
                                <input className="form-container-input" type="text" id="City" name="City" required placeholder="Town or City"></input>
                            </div>
                        </div>
                    </div>
                </div>

                <div className='row p-4'>
                    <div className='col-12 py-4 form-container'>
                        <div className='row p-0'>
                            <div className='col-12'>
                                <h5>Rental Dates</h5>
                            </div>
                            <div className='col-8 pr-4'>
                                <p><highlited-gray>Select available dates from the calendar</highlited-gray></p>
                            </div>
                            <div className='col-4 pl-4 t-e'>
                                <p><highlited-gray>Step 2 of 4</highlited-gray></p>
                            </div>
                            <div className='col-12'>
                                <div className='calendar-card'>
                                    <div className='calendar-controls'>
                                        <button type='button' className='calendar-nav' onClick={handlePrevMonth} disabled={currentMonth.getFullYear() === today.getFullYear() && currentMonth.getMonth() === today.getMonth()}>&lt;</button>
                                        <span className='calendar-month'>
                                            {currentMonth.toLocaleString('default', { month: 'long', year: 'numeric' })}
                                        </span>
                                        <button type='button' className='calendar-nav' onClick={handleNextMonth}>&gt;</button>
                                </div>
                                    <div className='calendar-grid calendar-weekdays'>
                                        {weekDays.map((day) => (
                                            <div key={day} className='calendar-weekday'>{day}</div>
                                        ))}
                            </div>
                                    <div className='calendar-grid'>
                                        {calendarDays.map((item, index) => {
                                            if (!item) {
                                                return <div key={index} className='calendar-day calendar-day--empty'></div>;
                                            }
                                            const classNames = [
                                                'calendar-day',
                                                item.disabled ? 'calendar-day--disabled' : '',
                                                item.booked ? 'calendar-day--booked' : '',
                                                isSelectedDate(item) ? 'calendar-day--selected' : '',
                                                !isSelectedDate(item) && isInRange(item) ? 'calendar-day--range' : '',
                                            ].join(' ').trim();
                                            return (
                                                <button
                                                    type='button'
                                                    key={`${item.day}-${index}`}
                                                    className={classNames}
                                                    onClick={() => handleDayClick(item)}
                                                    disabled={item.disabled}
                                                >
                                                    {item.day}
                                                </button>
                                            );
                                        })}
                            </div>
                                    <div className='calendar-selection mt-3'>
                                        <p className='mb-1'><strong>Pick-Up:</strong> {selectedStartDate ? formatDate(selectedStartDate) : 'Not selected'}</p>
                                        <p className='mb-0'><strong>Drop-Off:</strong> {selectedEndDate ? formatDate(selectedEndDate) : 'Not selected'}</p>
                                    </div>
                                </div>
                                <input type="hidden" name="DatePick" value={selectedStartDate ? formatDate(selectedStartDate) : ''} />
                                <input type="hidden" name="DateDrop" value={selectedEndDate ? formatDate(selectedEndDate) : ''} />
                            </div>
                        </div>
                    </div>
                </div>


                <div className='row p-4'>
                    <div className='col-12 py-4 form-container'>
                        <div className='row p-0'>
                            <div className='col-12'>
                                <h5>Payment Method</h5>
                            </div>
                            <div className='col-8 pr-4'>
                                <p><highlited-gray>Please enter your payment method</highlited-gray></p>
                            </div>
                            <div className='col-4 pl-4 t-e'>
                                <p><highlited-gray>Step 3 of 4</highlited-gray></p>
                            </div>
                            <div className="col-12 pt-4">
                                <div className="row credit-card p-4">
                                    <div className="col-12 p-0 mb-4">
                                        <div className="row credit-card-h">
                                            <img src="/images/pick-up-icon.png" className="mr-3"></img>
                                            <h6 className="m-0 mt-1">Credit Card</h6>
                                            <img className="ml-auto" src="/images/Visa.png"></img>
                                        </div>
                                    </div>


                                    <div className='col-6 pr-4'>
                                        <h6>Card Number</h6>
                                        <input
                                            className="credit-card-input"
                                            type="tel"
                                            id="CardNumber"
                                            name="CardNumber"
                                            placeholder="Card Number"
                                            onInput={handleCardChange}
                                        ></input>
                                    </div>
                                    <div className='col-6 pb-4 pl-4'>
                                        <h6>Expration Date</h6>
                                        <input
                                            className="credit-card-input"
                                            type="text"
                                            id="CardDate"
                                            name="CardDate"
                                            placeholder="DD/MM/YY"
                                            onInput={handleExpirationDateChange}
                                        ></input>
                                    </div>
                                    <div className='col-6 pr-4'>
                                        <h6>Card Holder</h6>
                                        <input
                                            className="credit-card-input"
                                            type="text"
                                            id="CardHolder"
                                            name="CardHolder"
                                            placeholder="Card holder"
                                            onInput={handleNameChange}
                                        ></input>
                                    </div>
                                    <div className='col-6 pl-4'>
                                        <h6>CVC</h6>
                                        <input
                                            className="credit-card-input"
                                            type="text"
                                            id="CardCvc"
                                            name="CardCvc"
                                            placeholder="CVC"
                                            onInput={handleCVCChange}
                                        ></input>
                                    </div>
                                </div>

                                {/* <div className="col-12 my-4 p-0">
                                    <div className="row credit-card-h2">
                                        <input type="radio" id="PayPal" name="PayPal" value="PayPal"></input>
                                        <label for="PayPal">PayPal</label>
                                        <img className="ml-auto" src="/images/PayPal.png"></img>
                                    </div>
                            </div>
                            <div className="col-12 my-4 p-0">
                                    <div className="row credit-card-h2">
                                        <input type="radio" id="Bitcoin" name="Bitcoin" value="Bitcoin"></input>
                                        <label for="Bitcoin">Bitcoin</label>
                                        <img className="ml-auto" src="/images/Bitcoin.png"></img>
                                    </div>
                            </div> */}
                            </div>
                        </div>
                    </div>
                </div>


                <div className='row p-4'>
                    <div className='col-12 py-4 form-container'>
                        <div className='row p-0'>
                            <div className='col-12'>
                                <h5>Confirmation</h5>
                            </div>
                            <div className='col-8 pr-4'>
                                <p><highlited-gray>We are getting to the end. Just few clicks and your rental is ready!</highlited-gray></p>
                            </div>
                            <div className='col-4 pl-4 t-e'>
                                <p><highlited-gray>Step 4 of 4</highlited-gray></p>
                            </div>
                            <div className="col-12">
                                <div className="col-12 my-4 p-0">
                                    <div className="row credit-card-h2">
                                        <input type="checkbox" id="SendNews" name="SendNews" value="true"></input>
                                        <label for="sendNews">I agree with sending an Marketing and newsletter emails. No spam, promissed!</label>
                                    </div>
                                </div>
                                <div className="col-12 my-4 p-0">
                                    <div className="row credit-card-h2">
                                        <input type="checkbox" id="PrivacyPolicy" name="PrivacyPolicy" value="true"></input>
                                        <label for="privacyPolicy">I agree with our terms and conditions and privacy policy.</label>
                                    </div>
                                </div>
                            </div>

                            <div className="col-12 pb-3">
                                <button type="submit" className="PayButton" id="paybtn" name="paybtn" disabled={submitting}>
                                    {submitting ? 'Processing...' : 'Rent Now'}
                                </button>
                            </div>
                            <div className="col-12 pb-3">
                                <img src="/images/shield.png" alt="" />
                            </div>
                            <div className="col-12">
                                <h6 className="m-0">All your data are safe</h6>
                                <p><highlited-gray>We are using the most advanced security to provide you the best experience ever.</highlited-gray></p>
                            </div>
                        </div>
                    </div>
                </div>
            </form>

        </div>
    );
};

export default PayForm;
    