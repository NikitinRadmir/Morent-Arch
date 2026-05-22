import React, { useCallback, useContext, useEffect, useMemo, useState } from 'react';
import toast from 'react-hot-toast';
import { useNavigate } from 'react-router-dom';
import { AuthContext, API_BASE_URL } from '../context/AuthContext';
import { bankApi } from '../api/bankApi';
import { formatUsd } from '../utils/formatMoney';

const PayForm = ({ car, setTotalAmount }) => {
    const { isAuthenticated, authRequest, user } = useContext(AuthContext);
    const navigate = useNavigate();
    const [bankProfile, setBankProfile] = useState(null);
    const [bankLoading, setBankLoading] = useState(false);
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
    const isDateInBookedRange = useCallback((date) => {
        if (!bookedDateStrings || bookedDateStrings.length === 0) return false;
        return bookedDateStrings.includes(normalizeDateToISO(date));
    }, [bookedDateStrings]);

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
    }, [currentMonth, today, isDateInBookedRange]);

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
    const calculateTotalAmount = useCallback((datePick, dateDrop) => {
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
    }, [car.price, setTotalAmount]);

    useEffect(() => {
        if (selectedStartDate && selectedEndDate) {
            calculateTotalAmount(formatDate(selectedStartDate), formatDate(selectedEndDate));
        }
    }, [selectedStartDate, selectedEndDate, calculateTotalAmount]);

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

    useEffect(() => {
        if (!isAuthenticated) {
            setBankProfile(null);
            return;
        }
        const loadBank = async () => {
            try {
                setBankLoading(true);
                await bankApi.syncSession();
                const profile = await bankApi.profile();
                setBankProfile(profile);
            } catch (error) {
                console.error('Failed to load bank profile for payment', error);
                setBankProfile(null);
            } finally {
                setBankLoading(false);
            }
        };
        loadBank();
    }, [isAuthenticated, user?.id]);

    const billingName = user?.nickname || user?.name || '—';
    const billingEmail = user?.email || '—';


    const mapRentalError = (message) => {
        const text = (message || '').toLowerCase();
        if (text.includes('insufficient') || text.includes('недостаточно средств')) {
            return 'Недостаточно средств на банковском счёте. Пополните карту в Morent Bank.';
        }
        if (text.includes('bank session') || text.includes('сессия банка') || text.includes('morent bank')) {
            return 'Откройте Morent Bank и обновите страницу оплаты.';
        }
        if (text.includes('уже забронирован') || text.includes('already booked')) {
            return 'Автомобиль уже забронирован на выбранные даты.';
        }
        if (text.includes('price') || text.includes('сумма аренды')) {
            return 'Сумма аренды устарела. Перевыберите даты.';
        }
        if (text.includes('bank') && text.includes('недоступен')) {
            return 'Банковский сервис временно недоступен. Попробуйте позже.';
        }
        return message || 'Не удалось оформить аренду';
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

        if (!bankProfile?.cardNumber) {
            addNotification('Bank card is not available. Open Morent Bank first.');
            toast.error('Банковская карта недоступна. Откройте Morent Bank.');
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

        const price = Number(localTotalAmount);
        if (bankProfile.balance != null && price > Number(bankProfile.balance)) {
            const msg = 'Недостаточно средств на банковском счёте';
            addNotification(msg);
            toast.error(msg);
            return;
        }

        try {
            setSubmitting(true);
            await bankApi.syncSession();
            const freshProfile = await bankApi.profile();
            setBankProfile(freshProfile);
            if (freshProfile?.balance != null && price > Number(freshProfile.balance)) {
                throw new Error('недостаточно средств на банковском счёте');
            }

            const rentalPayload = {
                carId: car.id,
                startDate: normalizeDateToISO(selectedStartDate),
                endDate: normalizeDateToISO(selectedEndDate),
                totalPrice: price,
            };
            const rentalResponse = await authRequest('/rentals', {
                method: 'POST',
                body: JSON.stringify(rentalPayload),
            });

            try {
                const afterPay = await bankApi.profile();
                setBankProfile(afterPay);
            } catch {
                // ignore profile refresh errors
            }

            toast.success('Бронирование успешно оформлено');
            navigate('/rental-success', {
                state: {
                    car,
                    rental: rentalResponse,
                },
            });
        } catch (error) {
            const msg = mapRentalError(error.message);
            setSubmitError(msg);
            addNotification(msg);
            toast.error(msg);
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
                    <div className='col-12 py-4 form-container'>
                        <div className='row p-0'>
                            <div className='col-12'>
                                <h5>Rental Dates</h5>
                            </div>
                            <div className='col-8 pr-4'>
                                <p><highlited-gray>Select available dates from the calendar</highlited-gray></p>
                            </div>
                            <div className='col-4 pl-4 t-e'>
                                <p><highlited-gray>Step 1 of 3</highlited-gray></p>
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
                                <h5>Billing &amp; Payment</h5>
                            </div>
                            <div className='col-8 pr-4'>
                                <p><highlited-gray>Данные подставляются из вашего профиля и Morent Bank</highlited-gray></p>
                            </div>
                            <div className='col-4 pl-4 t-e'>
                                <p><highlited-gray>Step 2 of 3</highlited-gray></p>
                            </div>
                            <div className="col-12 pt-3">
                                <div className="row pay-billing-summary mb-3">
                                    <div className="col-md-6 mb-2">
                                        <h6 className="mb-1">Name</h6>
                                        <p className="pay-readonly-field mb-0">{billingName}</p>
                                    </div>
                                    <div className="col-md-6 mb-2">
                                        <h6 className="mb-1">Email</h6>
                                        <p className="pay-readonly-field mb-0">{billingEmail}</p>
                                    </div>
                                </div>
                            </div>
                            <div className="col-12 pt-2">
                                <div className="row credit-card p-4 pay-card-readonly">
                                    <div className="col-12 p-0 mb-4">
                                        <div className="row credit-card-h">
                                            <img src="/images/pick-up-icon.png" className="mr-3" alt="" />
                                            <h6 className="m-0 mt-1">Morent Bank Card</h6>
                                            <img className="ml-auto" src="/images/Visa.png" alt="Visa" />
                                        </div>
                                    </div>
                                    {bankLoading ? (
                                        <div className="col-12">
                                            <p className="mb-0 text-muted">Загрузка карты…</p>
                                        </div>
                                    ) : bankProfile?.cardNumber ? (
                                        <>
                                            <div className="col-12 mb-3">
                                                <h6>Баланс счёта</h6>
                                                <p className={`pay-readonly-field mb-0 ${Number(bankProfile.balance) < Number(localTotalAmount) ? 'pay-balance--low' : ''}`}>
                                                    {formatUsd(bankProfile.balance)}
                                                    {Number(localTotalAmount) > 0 && (
                                                        <span className="pay-balance-hint">
                                                            {' '}
                                                            (к оплате: {formatUsd(localTotalAmount)})
                                                        </span>
                                                    )}
                                                </p>
                                            </div>
                                            <div className="col-12 mb-3">
                                                <h6>Card Number</h6>
                                                <p className="pay-readonly-field pay-card-number mb-0">{bankProfile.cardNumber}</p>
                                            </div>
                                            <div className="col-md-4 mb-2">
                                                <h6>Expiration</h6>
                                                <p className="pay-readonly-field mb-0">{bankProfile.expDate}</p>
                                            </div>
                                            <div className="col-md-4 mb-2">
                                                <h6>Card Holder</h6>
                                                <p className="pay-readonly-field mb-0">{bankProfile.cardHolder}</p>
                                            </div>
                                            <div className="col-md-4 mb-2">
                                                <h6>CVV</h6>
                                                <p className="pay-readonly-field mb-0">{bankProfile.cvv}</p>
                                            </div>
                                        </>
                                    ) : (
                                        <div className="col-12">
                                            <p className="mb-0 text-danger">Карта не найдена. Откройте <a href="/bank">Morent Bank</a>, чтобы выпустить виртуальную карту.</p>
                                        </div>
                                    )}
                                </div>
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
                                <p><highlited-gray>Step 3 of 3</highlited-gray></p>
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
