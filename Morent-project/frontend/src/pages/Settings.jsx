import React, { useContext, useEffect, useState } from 'react';
import { AuthContext, API_BASE_URL } from '../context/AuthContext';
import { useNavigate } from 'react-router-dom';
import NewPasswordField from '../components/NewPasswordField';
import { usePasswordField } from '../hooks/usePasswordField';

const uploadImage = async (file, setStatus) => {
  if (!file) return '';
  setStatus && setStatus('uploading');
  const formData = new FormData();
  formData.append('file', file);
  try {
    const res = await fetch(`${API_BASE_URL}/media/upload`, {
    method: 'POST',
    body: formData,
      credentials: 'include',
  });
  if (!res.ok) {
      setStatus && setStatus('error');
      const txt = await res.text();
      throw new Error(txt || 'Ошибка загрузки');
    }
    const data = await res.json();
    setStatus && setStatus('success');
    return data.url;
  } catch (err) {
    setStatus && setStatus('error');
    throw err;
  }
};

const Settings = () => {
    const { isAuthenticated, user, fetchProfile, updateProfile, changePassword } = useContext(AuthContext);
    const [activeTab, setActiveTab] = useState('avatar');
    const [avatarUrl, setAvatarUrl] = useState('');
    const [nickname, setNickname] = useState('');
    const [position, setPosition] = useState('');
    const [oldPassword, setOldPassword] = useState('');
    const [newPassword, setNewPassword] = useState('');
    const [confirmPassword, setConfirmPassword] = useState('');
    const [loading, setLoading] = useState(false);
    const [message, setMessage] = useState('');
    const [error, setError] = useState('');
    const [avatarUploadStatus, setAvatarUploadStatus] = useState('idle');
    const navigate = useNavigate();
    const newPasswordField = usePasswordField(newPassword);

    useEffect(() => {
        if (!isAuthenticated) {
            navigate('/sign-in');
            return;
        }
        const init = async () => {
            try {
                const profile = await fetchProfile();
                if (profile) {
                    setAvatarUrl(profile.avatarUrl || '');
                    setNickname(profile.nickname || profile.name || '');
                    setPosition(profile.position || '');
                }
            } catch {
                // ignore
            }
        };
        init();
    }, [isAuthenticated, fetchProfile, navigate]);

    const handleAvatarSave = async () => {
        setLoading(true);
        setMessage('');
        setError('');
        try {
            await updateProfile({ avatarUrl });
            setMessage('Avatar updated');
        } catch (e) {
            setError(e.message || 'Failed to update avatar');
        } finally {
            setLoading(false);
        }
    };

    const handleProfileSave = async () => {
        setLoading(true);
        setMessage('');
        setError('');
        try {
            await updateProfile({ nickname, position });
            setMessage('Profile updated');
        } catch (e) {
            setError(e.message || 'Failed to update profile');
        } finally {
            setLoading(false);
        }
    };

    const handlePasswordSave = async () => {
        setLoading(true);
        setMessage('');
        setError('');
        try {
            if (!oldPassword.trim()) {
                setError('Введите текущий пароль');
                setLoading(false);
                return;
            }
            if (newPassword !== confirmPassword) {
                setError('Новые пароли не совпадают');
                setLoading(false);
                return;
            }
            if (!newPasswordField.isValid) {
                setError('Новый пароль не соответствует требованиям безопасности');
                setLoading(false);
                return;
            }
            if (newPasswordField.checking) {
                setLoading(false);
                return;
            }
            await changePassword({ oldPassword, newPassword });
            setMessage('Пароль изменён');
            setOldPassword('');
            setNewPassword('');
            setConfirmPassword('');
            newPasswordField.resetGeneratorState();
        } catch (e) {
            setError(e.message || 'Не удалось сменить пароль');
        } finally {
            setLoading(false);
        }
    };

    const passwordsMatch =
        confirmPassword.length > 0 && newPassword === confirmPassword;
    const canChangePassword =
        !loading &&
        !newPasswordField.checking &&
        newPasswordField.isValid &&
        passwordsMatch &&
        oldPassword.trim().length > 0;

    if (!isAuthenticated) {
        return null;
    }

    return (
        <div className="background-gray py-4">
            <div className="container">
                <h2 className="mb-3">Profile settings</h2>
                <p className="highlited-gray mb-4">Manage your avatar, profile and password.</p>

                <div className="settings-tabs mb-4">
                    <button
                        type="button"
                        className={`settings-tab ${activeTab === 'avatar' ? 'settings-tab--active' : ''}`}
                        onClick={() => setActiveTab('avatar')}
                    >
                        Avatar
                    </button>
                    <button
                        type="button"
                        className={`settings-tab ${activeTab === 'profile' ? 'settings-tab--active' : ''}`}
                        onClick={() => setActiveTab('profile')}
                    >
                        Profile
                    </button>
                    <button
                        type="button"
                        className={`settings-tab ${activeTab === 'password' ? 'settings-tab--active' : ''}`}
                        onClick={() => setActiveTab('password')}
                    >
                        Password
                    </button>
                </div>

                {message && <div className="auth-alert auth-alert--success mb-3">{message}</div>}
                {error && <div className="auth-alert auth-alert--error mb-3">{error}</div>}

                <div className="settings-card">
                    {activeTab === 'avatar' && (
                        <div>
                            <h4 className="mb-3">Avatar</h4>
                            <div className="settings-avatar-preview mb-3">
                                <img src={avatarUrl || 'https://i.pravatar.cc/120?img=12'} alt="Avatar preview" />
                            </div>
                            {/* file upload */}
                            <div style={{display:'flex',alignItems:'center',gap:16,marginBottom:6}}>
                                <input type="file" accept="image/*" style={{width:180}} onChange={async e => {
                                    if(e.target.files?.[0]) {
                                        try {
                                            setAvatarUploadStatus('uploading');
                                            const url = await uploadImage(e.target.files[0], setAvatarUploadStatus);
                                            setAvatarUrl(url);
                                        } catch (err) {
                                            alert("Ошибка загрузки аватара: " + err.message);
                                        }
                                    }
                                }}/>
                                {avatarUploadStatus==='uploading' && <span>Загрузка...</span>}
                            </div>
                            <label className="settings-label">
                                Avatar URL
                                <input
                                    type="text"
                                    className="form-container-input mt-2"
                                    value={avatarUrl}
                                    onChange={(e) => setAvatarUrl(e.target.value)}
                                    placeholder="https://..."
                                />
                            </label>
                            <button
                                type="button"
                                className="auth-submit mt-3"
                                onClick={handleAvatarSave}
                                disabled={loading}
                            >
                                {loading ? 'Saving...' : 'Save avatar'}
                            </button>
                        </div>
                    )}

                    {activeTab === 'profile' && (
                        <div>
                            <h4 className="mb-3">Profile</h4>
                            <label className="settings-label">
                                Nickname
                                <input
                                    type="text"
                                    className="form-container-input mt-2"
                                    value={nickname}
                                    onChange={(e) => setNickname(e.target.value)}
                                    placeholder={user?.name || ''}
                                />
                            </label>
                            <label className="settings-label mt-3">
                                Position
                                <input
                                    type="text"
                                    className="form-container-input mt-2"
                                    value={position}
                                    onChange={(e) => setPosition(e.target.value)}
                                    placeholder="e.g. Product Manager"
                                />
                            </label>
                            <button
                                type="button"
                                className="auth-submit mt-3"
                                onClick={handleProfileSave}
                                disabled={loading}
                            >
                                {loading ? 'Saving...' : 'Save profile'}
                            </button>
                        </div>
                    )}

                    {activeTab === 'password' && (
                        <div>
                            <h4 className="mb-3">Смена пароля</h4>
                            <label className="settings-label">
                                Текущий пароль
                                <input
                                    type="password"
                                    className="form-container-input mt-2"
                                    value={oldPassword}
                                    onChange={(e) => setOldPassword(e.target.value)}
                                    autoComplete="current-password"
                                />
                            </label>
                            <div className="mt-3">
                                <NewPasswordField
                                    label="Новый пароль"
                                    value={newPassword}
                                    onChange={(value) => {
                                        setNewPassword(value);
                                        newPasswordField.resetGeneratorState();
                                    }}
                                    onGenerated={(value) => {
                                        setNewPassword(value);
                                        setConfirmPassword(value);
                                    }}
                                    disabled={loading}
                                    requirementsId="settings-password-requirements"
                                    labelWrapper="settings"
                                    inputClassName="form-container-input"
                                    placeholder="Введите новый пароль"
                                    passwordField={newPasswordField}
                                />
                            </div>
                            <label className="settings-label mt-3">
                                Подтвердите новый пароль
                                <input
                                    type="password"
                                    className="form-container-input mt-2"
                                    value={confirmPassword}
                                    onChange={(e) => setConfirmPassword(e.target.value)}
                                    autoComplete="new-password"
                                    aria-invalid={confirmPassword.length > 0 && !passwordsMatch}
                                />
                                {confirmPassword.length > 0 && !passwordsMatch && (
                                    <p className="password-hint password-hint--error">Пароли не совпадают</p>
                                )}
                            </label>
                            <button
                                type="button"
                                className="auth-submit mt-3"
                                onClick={handlePasswordSave}
                                disabled={!canChangePassword}
                            >
                                {loading ? 'Сохранение…' : 'Сменить пароль'}
                            </button>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
};

export default Settings;




