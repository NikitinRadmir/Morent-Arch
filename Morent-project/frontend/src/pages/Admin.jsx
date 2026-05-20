import React, { useContext, useEffect, useMemo, useState } from 'react';
import { AuthContext, API_BASE_URL } from '../context/AuthContext';
import NotFound from './NotFound';

const API_BASE = API_BASE_URL;
const estimateRentalPrice = (msrp) => {
    const value = Number(msrp);
    if (!Number.isFinite(value) || value <= 0) return 0;
    return value * 0.01 / 2;
};

const uploadImage = async (file) => {
    if (!file) return '';
    const formData = new FormData();
    formData.append('file', file);
    const res = await fetch(`${API_BASE}/Admin/Media/Upload`, {
        method: 'POST',
        body: formData,
        credentials: 'include',
    });
    if (!res.ok) {
        const txt = await res.text();
        throw new Error(txt || 'Ошибка загрузки');
    }
    const data = await res.json();
    return data.url;
};

const tabs = [
    { id: 'cars', label: 'Машины' },
    { id: 'users', label: 'Пользователи' },
    { id: 'rentals', label: 'Аренды' },
    { id: 'comments', label: 'Комментарии' },
    { id: 'favorites', label: 'Избранное' },
    { id: 'logs', label: 'История' },
];

const initialCar = {
    name: '',
    type: '',
    capacity: 4,
    price: 0,
    fuel: 0,
    transmission: 'Automatic',
    imgSrc: '',
    description: '',
};

const initialUser = {
    name: '',
    email: '',
    password: '',
    position: '',
    avatarUrl: '',
    role: 'user',
};

const initialComment = {
    id: 0,
    description: '',
    rating: 5,
};

const Modal = ({ title, children, onClose }) => (
    <div className="admin-modal-backdrop" role="dialog" aria-modal="true">
        <div className="admin-modal">
            <div className="admin-modal__header">
                <h3>{title}</h3>
                <button type="button" className="admin-btn admin-btn--ghost" onClick={onClose}>×</button>
            </div>
            <div className="admin-modal__body">{children}</div>
        </div>
    </div>
);

const Admin = () => {
    const { user, fetchProfile } = useContext(AuthContext);
    const [activeTab, setActiveTab] = useState('cars');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const [cars, setCars] = useState([]);
    const [users, setUsers] = useState([]);
    const [rentals, setRentals] = useState([]);
    const [comments, setComments] = useState([]);
    const [favorites, setFavorites] = useState([]);
    const [logs, setLogs] = useState([]);
    const [aggregatorQuery, setAggregatorQuery] = useState('');
    const [aggregatorCars, setAggregatorCars] = useState([]);
    const [aggregatorLoading, setAggregatorLoading] = useState(false);
    const [aggregatorStatus, setAggregatorStatus] = useState('');
    const [importingTrimId, setImportingTrimId] = useState(null);
    const [modal, setModal] = useState(null); // { type, form, data }
    const [uploadingCarImg, setUploadingCarImg] = useState(false);
    const [uploadingUserAvatar, setUploadingUserAvatar] = useState(false);

    const jsonHeaders = useMemo(() => ({
        'Content-Type': 'application/json',
    }), []);

    useEffect(() => {
        // Проверка роли при загрузке
        const userRole = user?.role || user?.Role;
        if (userRole !== 'admin') {
            // Не админ - показываем 404
            return;
        }
        loadTab(activeTab);
    }, [activeTab, user]);

    // Если пользователь не админ - показываем 404
    const userRole = user?.role || user?.Role;
    if (!userRole || userRole !== 'admin') {
        console.log('Admin access denied. User role:', userRole, 'User object:', user);
        return <NotFound />;
    }

    const fetchJson = async (url, options = {}) => {
        // cookie-сессия отправляется автоматически
        const headers = {
            ...(options.headers || {}),
        };
        const resp = await fetch(url, { ...options, headers, credentials: 'include' });
        if (!resp.ok) {
            const text = await resp.text();
            throw new Error(text || resp.statusText);
        }
        if (resp.status === 204) return null;
        const txt = await resp.text();
        return txt ? JSON.parse(txt) : null;
    };

    const loadTab = async (tab) => {
        setLoading(true);
        setError('');
        try {
            switch (tab) {
                case 'cars':
                    setCars(await fetchJson(`${API_BASE}/cars`));
                    break;
                case 'users':
                    setUsers(await fetchJson(`${API_BASE}/Admin/Users`));
                    break;
                case 'rentals':
                    setRentals(await fetchJson(`${API_BASE}/Admin/Rentals`));
                    break;
                case 'comments':
                    setComments(await fetchJson(`${API_BASE}/Admin/Comments`));
                    break;
                case 'favorites':
                    setFavorites(await fetchJson(`${API_BASE}/Admin/Favorites`));
                    break;
                case 'logs':
                    setLogs(await fetchJson(`${API_BASE}/Admin/Logs`));
                    break;
                default:
                    break;
            }
        } catch (e) {
            setError(e.message || 'Не удалось загрузить данные');
        } finally {
            setLoading(false);
        }
    };

    const handleCreateCar = async (payload) => {
        await fetchJson(`${API_BASE}/Admin/Cars`, {
            method: 'POST',
            headers: jsonHeaders,
            body: JSON.stringify(payload),
        });
        await loadTab('cars');
    };

    const handleUpdateCar = async (payload) => {
        await fetchJson(`${API_BASE}/Admin/Cars/${payload.id}`, {
            method: 'PUT',
            headers: jsonHeaders,
            body: JSON.stringify(payload),
        });
        await loadTab('cars');
    };

    const handleDeleteCar = async (id) => {
        await fetchJson(`${API_BASE}/Admin/Cars/${id}`, { method: 'DELETE', headers: jsonHeaders });
        await loadTab('cars');
    };

    const loadAggregatorCars = async () => {
        const query = aggregatorQuery.trim();
        if (!query) {
            setAggregatorStatus('Введите запрос, например "Golf" или "Toyota Camry"');
            setAggregatorCars([]);
            return;
        }

        setAggregatorLoading(true);
        setAggregatorStatus('');
        try {
            const list = await fetchJson(`${API_BASE}/Admin/Aggregator/Cars?q=${encodeURIComponent(query)}`);
            setAggregatorCars(list?.cars || []);
            if (!list?.cars?.length) {
                setAggregatorStatus('По вашему запросу в агрегаторе ничего не найдено');
            }
        } catch (e) {
            const message = e.message || 'Ошибка запроса к агрегатору';
            if (message.toLowerCase().includes('агрегатора') || message.includes('503')) {
                setAggregatorStatus('Микросервис агрегатора сейчас недоступен');
            } else {
                setAggregatorStatus(message);
            }
            setAggregatorCars([]);
        } finally {
            setAggregatorLoading(false);
        }
    };

    const importAggregatorCar = async (trim) => {
        setImportingTrimId(trim.id);
        setAggregatorStatus('');
        try {
            await fetchJson(`${API_BASE}/Admin/Aggregator/Import`, {
                method: 'POST',
                headers: jsonHeaders,
                body: JSON.stringify({ trim }),
            });
            setAggregatorStatus(`Машина "${trim.make} ${trim.model} ${trim.trim}" добавлена в БД`);
            await loadTab('cars');
        } catch (e) {
            setAggregatorStatus(e.message || 'Не удалось импортировать машину');
        } finally {
            setImportingTrimId(null);
        }
    };

    const handleCreateUser = async (payload) => {
        const registerResp = await fetchJson(`${API_BASE}/auth/register`, {
            method: 'POST',
            headers: jsonHeaders,
            body: JSON.stringify({ name: payload.name, email: payload.email, password: payload.password }),
        });
        const created = registerResp?.user;
        if (!created) {
            await loadTab('users');
            return;
        }
        if (payload.position || payload.avatarUrl || payload.role) {
            await fetchJson(`${API_BASE}/Admin/Users`, {
                method: 'PUT',
                headers: jsonHeaders,
                body: JSON.stringify({
                    id: created.id,
                    position: payload.position,
                    avatarUrl: payload.avatarUrl,
                    role: payload.role,
                    name: created.name,
                    email: created.email,
                }),
            });
        }
        await loadTab('users');
    };

    const handleUpdateUser = async (payload) => {
        const { nickname: _nickname, ...rest } = payload; // никнейм убрали из формы, не отправляем
        await fetchJson(`${API_BASE}/Admin/Users`, {
            method: 'PUT',
            headers: jsonHeaders,
            body: JSON.stringify(rest),
        });
        // Если обновляется текущий пользователь - обновляем его данные в AuthContext
        const currentUserId = user?.ID ?? user?.id;
        const updatedUserId = rest.ID ?? rest.id;
        if (currentUserId && updatedUserId && currentUserId === updatedUserId) {
            try {
                await fetchProfile();
            } catch (e) {
                console.error('Failed to refresh profile after update', e);
            }
        }
        await loadTab('users');
    };

    const handleDeleteUser = async (id, user) => {
        if (!id || typeof id !== 'number') {
            setError('Ошибка удаления: Не передан ID пользователя! ' + String(id));
            console.error('Bad delete user!', id, user);
            return;
        }
        try {
            const url = `${API_BASE}/Admin/Users/${id}`;
            console.log('Удаление пользователя', id, user, url);
            await fetchJson(url, { method: 'DELETE', headers: jsonHeaders });
            await loadTab('users');
        } catch (e) {
            let msg = e && e.message ? e.message : String(e);
            setError('Ошибка удаления: ' + msg);
        }
    };


    const handleDeleteRental = async (id) => {
        await fetchJson(`${API_BASE}/Admin/Rentals/${id}`, { method: 'DELETE', headers: jsonHeaders });
        await loadTab('rentals');
    };

    const handleUpdateComment = async (payload) => {
        await fetchJson(`${API_BASE}/Admin/Comments`, {
            method: 'PUT',
            headers: jsonHeaders,
            body: JSON.stringify(payload),
        });
        await loadTab('comments');
    };

    const handleDeleteComment = async (id) => {
        await fetchJson(`${API_BASE}/Admin/Comments/${id}`, { method: 'DELETE', headers: jsonHeaders });
        await loadTab('comments');
    };

    const handleDeleteFavorite = async (userId, carId) => {
        await fetchJson(`${API_BASE}/Admin/Favorites/${userId}/${carId}`, { method: 'DELETE', headers: jsonHeaders });
        await loadTab('favorites');
    };

    const openModal = (type, data = {}) => {
        if (type === 'car-create') {
            setModal({ type, form: { ...initialCar } });
            return;
        }
        if (type === 'car-edit') {
            setModal({ type, form: { ...initialCar, ...data }, data });
            return;
        }
        if (type === 'user-create') {
            setModal({ type, form: { ...initialUser } });
            return;
        }
        if (type === 'user-edit') {
            setModal({ type, form: { ...initialUser, ...data, password: '' }, data });
            return;
        }
        if (type === 'comment-edit') {
            setModal({ type, form: { ...initialComment, ...data }, data });
            return;
        }
        setModal({ type, data });
    };

    const updateModalForm = (patch) => {
        setModal((m) => (m ? { ...m, form: { ...m.form, ...patch } } : m));
    };

    const renderCars = () => (
        <div className="admin-card">
            <div className="admin-card__header">
                <h3>Машины</h3>
                <button className="admin-btn" onClick={() => openModal('car-create')}>Добавить</button>
            </div>
            <div className="admin-table-wrapper">
                <table className="admin-table">
                    <thead>
                        <tr>
                            <th>ID</th>
                            <th>Изображение</th>
                            <th>Название</th>
                            <th>Тип</th>
                            <th>Цена</th>
                            <th>Мест</th>
                            <th>Трансмиссия</th>
                            <th></th>
                        </tr>
                    </thead>
                    <tbody>
                        {cars.map((car) => (
                            <tr key={car.id}>
                                <td>{car.id}</td>
                                <td>
                                    {car.imgSrc ? (
                                        <img src={car.imgSrc} alt={car.name} className="admin-table__img" />
                                    ) : (
                                        <span className="admin-table__no-img">—</span>
                                    )}
                                </td>
                                <td>{car.name}</td>
                                <td>{car.type}</td>
                                <td>${car.price}</td>
                                <td>{car.capacity}</td>
                                <td>{car.transmission}</td>
                                <td className="admin-table__actions">
                                    <button className="admin-btn admin-btn--ghost" onClick={() => openModal('car-edit', car)}>Изменить</button>
                                    <button className="admin-btn admin-btn--danger" onClick={() => openModal('car-delete', car)}>Удалить</button>
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
            <div className="admin-aggregator">
                <h4>Машины из агрегатора</h4>
                <div className="admin-aggregator__controls">
                    <input
                        value={aggregatorQuery}
                        onChange={(e) => setAggregatorQuery(e.target.value)}
                        placeholder="Например: Golf, BMW X5, Toyota Camry"
                    />
                    <button className="admin-btn" onClick={loadAggregatorCars} disabled={aggregatorLoading}>
                        {aggregatorLoading ? 'Поиск...' : 'Найти в агрегаторе'}
                    </button>
                </div>
                {aggregatorStatus && <div className="admin-alert">{aggregatorStatus}</div>}
                {aggregatorCars.length > 0 && (
                    <div className="admin-table-wrapper">
                        <table className="admin-table">
                            <thead>
                                <tr>
                                    <th>ID</th>
                                    <th>Фото</th>
                                    <th>Год</th>
                                    <th>Марка</th>
                                    <th>Модель</th>
                                    <th>Комплектация</th>
                                    <th>Коробка</th>
                                    <th>Мест</th>
                                    <th>Расход (л/100км)</th>
                                    <th>Примерная цена аренды</th>
                                    <th></th>
                                </tr>
                            </thead>
                            <tbody>
                                {aggregatorCars.map((trim) => (
                                    <tr key={trim.id}>
                                        <td>{trim.id}</td>
                                        <td>
                                            {trim.imageUrl ? (
                                                <img src={trim.imageUrl} alt={`${trim.make} ${trim.model}`} className="admin-table__img" />
                                            ) : (
                                                <span className="admin-table__no-img">—</span>
                                            )}
                                        </td>
                                        <td>{trim.year}</td>
                                        <td>{trim.make}</td>
                                        <td>{trim.model}</td>
                                        <td>{trim.trim}</td>
                                        <td>{trim.transmission || '—'}</td>
                                        <td>{trim.seats || '—'}</td>
                                        <td>{trim.fuel ? trim.fuel.toFixed(1) : '—'}</td>
                                        <td>{trim.msrp ? `$${estimateRentalPrice(trim.msrp).toFixed(2)}` : '—'}</td>
                                        <td className="admin-table__actions">
                                            <button
                                                className="admin-btn"
                                                onClick={() => importAggregatorCar(trim)}
                                                disabled={importingTrimId === trim.id}
                                            >
                                                {importingTrimId === trim.id ? 'Добавление...' : 'Добавить в БД'}
                                            </button>
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                )}
            </div>
        </div>
    );

    const renderUsers = () => (
        <div className="admin-card">
            <div className="admin-card__header">
                <h3>Пользователи</h3>
                <button className="admin-btn" onClick={() => openModal('user-create')}>Добавить</button>
            </div>
            <div className="admin-table-wrapper">
                <table className="admin-table">
                    <thead>
                        <tr>
                            <th>ID</th>
                            <th>Аватар</th>
                            <th>Имя</th>
                            <th>Email</th>
                            <th>Ник</th>
                            <th>Роль</th>
                            <th></th>
                        </tr>
                    </thead>
                    <tbody>
                        {users.map((u) => (
                            <tr key={u.ID ?? u.id}>
                                <td>{u.id}</td>
                                <td>
                                    {u.avatarUrl ? (
                                        <img src={u.avatarUrl} alt={u.name} className="admin-table__img admin-table__img--avatar" />
                                    ) : (
                                        <span className="admin-table__no-img">—</span>
                                    )}
                                </td>
                                <td>{u.name}</td>
                                <td>{u.email}</td>
                                <td>{u.nickname}</td>
                                <td>{u.role}</td>
                                <td className="admin-table__actions">
                                    <button className="admin-btn admin-btn--ghost" onClick={() => openModal('user-edit', u)}>Изменить</button>
                                    <button className="admin-btn admin-btn--danger" onClick={() => openModal('user-delete', u)}>Удалить</button>
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        </div>
    );

    const renderRentals = () => (
        <div className="admin-card">
            <div className="admin-card__header">
                <h3>Аренды</h3>
            </div>
            <div className="admin-table-wrapper">
                <table className="admin-table">
                    <thead>
                        <tr>
                            <th>ID</th>
                            <th>User</th>
                            <th>Car</th>
                            <th>Даты</th>
                            <th>Цена</th>
                            <th></th>
                        </tr>
                    </thead>
                    <tbody>
                        {rentals.map((r) => {
                            const userId = r.userId ?? r.UserID;
                            const carId = r.carId ?? r.CarID;
                            const start = r.startDate ?? r.StartDate;
                            const end = r.endDate ?? r.EndDate;
                            const total = r.totalPrice ?? r.TotalPrice;
                            return (
                                <tr key={r.id ?? r.ID}>
                                    <td>{r.id ?? r.ID}</td>
                                    <td>{userId}</td>
                                    <td>{r.car?.name || carId}</td>
                                    <td>{start ? new Date(start).toLocaleDateString() : ''} — {end ? new Date(end).toLocaleDateString() : ''}</td>
                                    <td>${total}</td>
                                    <td className="admin-table__actions">
                                        <button className="admin-btn admin-btn--danger" onClick={() => openModal('rental-delete', r)}>Удалить</button>
                                    </td>
                                </tr>
                            );
                        })}
                    </tbody>
                </table>
            </div>
        </div>
    );

    const renderComments = () => (
        <div className="admin-card">
            <div className="admin-card__header">
                <h3>Комментарии</h3>
            </div>
            <div className="admin-table-wrapper">
                <table className="admin-table">
                    <thead>
                        <tr>
                            <th>ID</th>
                            <th>Car</th>
                            <th>Автор</th>
                            <th>Оценка</th>
                            <th>Текст</th>
                            <th></th>
                        </tr>
                    </thead>
                    <tbody>
                        {comments.map((c) => (
                            <tr key={c.id}>
                                <td>{c.id}</td>
                                <td>{c.carId}</td>
                                <td>{c.name}</td>
                                <td>{c.rating}</td>
                                <td className="admin-table__description">{c.description}</td>
                                <td className="admin-table__actions">
                                    <button className="admin-btn admin-btn--ghost" onClick={() => openModal('comment-edit', c)}>Изменить</button>
                                    <button className="admin-btn admin-btn--danger" onClick={() => openModal('comment-delete', c)}>Удалить</button>
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        </div>
    );

    const renderFavorites = () => (
        <div className="admin-card">
            <div className="admin-card__header">
                <h3>Избранное</h3>
            </div>
            <div className="admin-table-wrapper">
                <table className="admin-table">
                    <thead>
                        <tr>
                            <th>User</th>
                            <th>Car</th>
                            <th></th>
                        </tr>
                    </thead>
                    <tbody>
                        {favorites.map((f) => {
                            const userId = f.userId ?? f.UserID;
                            const carId = f.carId ?? f.CarID;
                            return (
                                <tr key={`${userId}-${carId}`}>
                                    <td>{userId}</td>
                                    <td>{f.car?.name || carId}</td>
                                    <td className="admin-table__actions">
                                        <button className="admin-btn admin-btn--danger" onClick={() => openModal('favorite-delete', { ...f, userId, carId })}>Удалить</button>
                                    </td>
                                </tr>
                            );
                        })}
                    </tbody>
                </table>
            </div>
        </div>
    );

    const renderLogs = () => (
        <div className="admin-card">
            <div className="admin-card__header">
                <h3>История изменений (сегодня)</h3>
            </div>
            <div className="admin-table-wrapper">
                <table className="admin-table">
                    <thead>
                        <tr>
                            <th>Время</th>
                            <th>Тип</th>
                            <th>Действие</th>
                            <th>UserID</th>
                            <th>ObjectID</th>
                            <th>Результат</th>
                            <th>Сообщение</th>
                        </tr>
                    </thead>
                    <tbody>
                        {logs.map((e, idx) => (
                            <tr key={idx}>
                                <td>{new Date(e.time).toLocaleTimeString()}</td>
                                <td>{e.type}</td>
                                <td>{e.action}</td>
                                <td>{e.userId ?? ''}</td>
                                <td>{e.objectId ?? ''}</td>
                                <td>{e.result}</td>
                                <td className="admin-table__description">{e.message}</td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        </div>
    );

    const renderContent = () => {
        if (loading) {
            return <div className="admin-loader">Загрузка...</div>;
        }
        switch (activeTab) {
            case 'cars':
                return renderCars();
            case 'users':
                return renderUsers();
            case 'rentals':
                return renderRentals();
            case 'comments':
                return renderComments();
            case 'favorites':
                return renderFavorites();
            case 'logs':
                return renderLogs();
            default:
                return null;
        }
    };

    const renderModal = () => {
        if (!modal) return null;
        const close = () => setModal(null);

        if (modal.type === 'car-create' || modal.type === 'car-edit') {
            const isEdit = modal.type === 'car-edit';
            const form = modal.form || { ...initialCar };
            return (
                <Modal title={isEdit ? 'Изменить машину' : 'Добавить машину'} onClose={close}>
                    <div className="admin-form">
                        <label>Название<input value={form.name} onChange={(e) => updateModalForm({ name: e.target.value })} /></label>
                        <label>Тип<input value={form.type} onChange={(e) => updateModalForm({ type: e.target.value })} /></label>
                        <label>Мест<input type="number" step="any" value={form.capacity} onChange={(e) => updateModalForm({ capacity: Number(e.target.value) })} /></label>
                        <label>Цена за день<input type="number" step="any" value={form.price} onChange={(e) => updateModalForm({ price: Number(e.target.value) })} /></label>
                        <label>Расход топлива<input type="number" step="any" value={form.fuel} onChange={(e) => updateModalForm({ fuel: Number(e.target.value) })} /></label>
                        <label>Трансмиссия<input value={form.transmission} onChange={(e) => updateModalForm({ transmission: e.target.value })} /></label>
                        <label>
                            Изображение
                            <div style={{ display: 'flex', alignItems: 'center', gap: '12px', marginTop: '6px' }}>
                                <input type="file" accept="image/*" style={{ width: '180px' }} onChange={async (e) => {
                                    if (e.target.files?.[0]) {
                                        try {
                                            setUploadingCarImg(true);
                                            const url = await uploadImage(e.target.files[0]);
                                            updateModalForm({ imgSrc: url });
                                        } catch (err) {
                                            setError('Ошибка загрузки изображения: ' + err.message);
                                        } finally {
                                            setUploadingCarImg(false);
                                        }
                                    }
                                }} />
                                {uploadingCarImg && <span>Загрузка...</span>}
                            </div>
                            {form.imgSrc && (
                                <div style={{ marginTop: '8px' }}>
                                    <img src={form.imgSrc} alt="Preview" style={{ maxWidth: '200px', maxHeight: '150px', borderRadius: '8px', border: '1px solid #e5e7eb' }} />
                                </div>
                            )}
                            <input type="text" value={form.imgSrc} onChange={(e) => updateModalForm({ imgSrc: e.target.value })} placeholder="Или введите URL" style={{ marginTop: '8px' }} />
                        </label>
                        <label>Описание<textarea value={form.description} onChange={(e) => updateModalForm({ description: e.target.value })} /></label>
                        <div className="admin-modal__footer">
                            <button className="admin-btn admin-btn--ghost" onClick={close}>Отмена</button>
                            <button
                                className="admin-btn"
                                onClick={async () => {
                                    try {
                                        if (isEdit) {
                                            await handleUpdateCar({ ...form, id: modal.data.id });
                                        } else {
                                            await handleCreateCar(form);
                                        }
                                        close();
                                    } catch (e) {
                                        setError(e.message);
                                    }
                                }}
                            >
                                Сохранить
                            </button>
                        </div>
                    </div>
                </Modal>
            );
        }

        if (modal.type === 'car-delete') {
            return (
                <Modal title="Удалить машину" onClose={close}>
                    <p>Удалить {modal.data.name}?</p>
                    <div className="admin-modal__footer">
                        <button className="admin-btn admin-btn--ghost" onClick={close}>Отмена</button>
                        <button className="admin-btn admin-btn--danger" onClick={async () => { await handleDeleteCar(modal.data.id); close(); }}>Удалить</button>
                    </div>
                </Modal>
            );
        }

        if (modal.type === 'user-create' || modal.type === 'user-edit') {
            const isEdit = modal.type === 'user-edit';
            const form = modal.form || { ...initialUser };
            return (
                <Modal title={isEdit ? 'Изменить пользователя' : 'Добавить пользователя'} onClose={close}>
                    <div className="admin-form">
                        <label>Имя<input value={form.name} onChange={(e) => updateModalForm({ name: e.target.value })} /></label>
                        <label>Email<input value={form.email} onChange={(e) => updateModalForm({ email: e.target.value })} /></label>
                        {!isEdit && (
                            <label>Пароль<input type="password" value={form.password} onChange={(e) => updateModalForm({ password: e.target.value })} /></label>
                        )}
                        <label>Должность<input value={form.position || ''} onChange={(e) => updateModalForm({ position: e.target.value })} /></label>
                        <label>
                            Аватар
                            <div style={{ display: 'flex', alignItems: 'center', gap: '12px', marginTop: '6px' }}>
                                <input type="file" accept="image/*" style={{ width: '180px' }} onChange={async (e) => {
                                    if (e.target.files?.[0]) {
                                        try {
                                            setUploadingUserAvatar(true);
                                            const url = await uploadImage(e.target.files[0]);
                                            updateModalForm({ avatarUrl: url });
                                        } catch (err) {
                                            setError('Ошибка загрузки аватара: ' + err.message);
                                        } finally {
                                            setUploadingUserAvatar(false);
                                        }
                                    }
                                }} />
                                {uploadingUserAvatar && <span>Загрузка...</span>}
                            </div>
                            {form.avatarUrl && (
                                <div style={{ marginTop: '8px' }}>
                                    <img src={form.avatarUrl} alt="Avatar preview" style={{ width: '80px', height: '80px', borderRadius: '50%', objectFit: 'cover', border: '1px solid #e5e7eb' }} />
                                </div>
                            )}
                            <input type="text" value={form.avatarUrl || ''} onChange={(e) => updateModalForm({ avatarUrl: e.target.value })} placeholder="Или введите URL" style={{ marginTop: '8px' }} />
                        </label>
                        <label>Роль<input value={form.role || ''} onChange={(e) => updateModalForm({ role: e.target.value })} /></label>
                        <div className="admin-modal__footer">
                            <button className="admin-btn admin-btn--ghost" onClick={close}>Отмена</button>
                            <button
                                className="admin-btn"
                                onClick={async () => {
                                    try {
                                        if (isEdit) {
                                            const userId = modal.data.ID ?? modal.data.id;
                                            await handleUpdateUser({ ...form, ID: userId, id: userId });
                                        } else {
                                            await handleCreateUser(form);
                                        }
                                        close();
                                    } catch (e) {
                                        setError(e.message);
                                    }
                                }}
                            >
                                Сохранить
                            </button>
                        </div>
                    </div>
                </Modal>
            );
        }

        if (modal.type === 'user-delete') {
            return (
                <Modal title="Удалить пользователя" onClose={close}>
                    <p>Удалить {modal.data.email}?</p>
                    <div className="admin-modal__footer">
                        <button className="admin-btn admin-btn--ghost" onClick={close}>Отмена</button>
                        <button className="admin-btn admin-btn--danger" onClick={async () => { await handleDeleteUser(modal.data.ID ?? modal.data.id, modal.data); close(); }}>Удалить</button>
                    </div>
                </Modal>
            );
        }

        if (modal.type === 'rental-delete') {
            return (
                <Modal title="Удалить аренду" onClose={close}>
                    <p>Удалить аренду #{modal.data.id ?? modal.data.ID}?</p>
                    <div className="admin-modal__footer">
                        <button className="admin-btn admin-btn--ghost" onClick={close}>Отмена</button>
                        <button className="admin-btn admin-btn--danger" onClick={async () => { await handleDeleteRental(modal.data.id ?? modal.data.ID); close(); }}>Удалить</button>
                    </div>
                </Modal>
            );
        }

        if (modal.type === 'comment-edit') {
            const form = modal.form || { ...initialComment };
            return (
                <Modal title="Изменить комментарий" onClose={close}>
                    <div className="admin-form">
                        <label>Оценка<input type="number" step="any" min="1" max="5" value={form.rating} onChange={(e) => updateModalForm({ rating: Number(e.target.value) })} /></label>
                        <label>Текст<textarea value={form.description} onChange={(e) => updateModalForm({ description: e.target.value })} /></label>
                        <div className="admin-modal__footer">
                            <button className="admin-btn admin-btn--ghost" onClick={close}>Отмена</button>
                            <button className="admin-btn" onClick={async () => { await handleUpdateComment(form); close(); }}>Сохранить</button>
                        </div>
                    </div>
                </Modal>
            );
        }

        if (modal.type === 'comment-delete') {
            return (
                <Modal title="Удалить комментарий" onClose={close}>
                    <p>Удалить комментарий #{modal.data.id}?</p>
                    <div className="admin-modal__footer">
                        <button className="admin-btn admin-btn--ghost" onClick={close}>Отмена</button>
                        <button className="admin-btn admin-btn--danger" onClick={async () => { await handleDeleteComment(modal.data.id); close(); }}>Удалить</button>
                    </div>
                </Modal>
            );
        }

        if (modal.type === 'favorite-delete') {
            return (
                <Modal title="Удалить из избранного" onClose={close}>
                    <p>Удалить car #{modal.data.carId} для user #{modal.data.userId}?</p>
                    <div className="admin-modal__footer">
                        <button className="admin-btn admin-btn--ghost" onClick={close}>Отмена</button>
                        <button className="admin-btn admin-btn--danger" onClick={async () => { await handleDeleteFavorite(modal.data.userId, modal.data.carId); close(); }}>Удалить</button>
                    </div>
                </Modal>
            );
        }

        return null;
    };

    return (
        <div className="admin-page">
            <div className="admin-header">
                <h2>Админ-панель</h2>
                <div className="admin-tabs">
                    {tabs.map((tab) => (
                        <button
                            key={tab.id}
                            className={`admin-tab ${activeTab === tab.id ? 'is-active' : ''}`}
                            onClick={() => setActiveTab(tab.id)}
                        >
                            {tab.label}
                        </button>
                    ))}
                </div>
            </div>

            {error && <div className="admin-alert">{error}</div>}

            {renderContent()}
            {renderModal()}
        </div>
    );
};

export default Admin;



