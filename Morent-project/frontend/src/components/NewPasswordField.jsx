import React from 'react';
import { PASSWORD_REQUIREMENTS } from '../utils/passwordFallback';

const NewPasswordField = ({
    label,
    value,
    onChange,
    onGenerated,
    disabled,
    requirementsId,
    labelWrapper = 'auth',
    inputClassName = '',
    placeholder = 'Новый пароль',
    passwordField,
}) => {
    const {
        checking,
        usedFallback,
        isInvalid,
        reqStatus,
        generating,
        generatorUnavailable,
        generateError,
        generate,
    } = passwordField;

    const handleGenerate = async () => {
        const result = await generate();
        if (result.ok && result.password) {
            onChange(result.password);
            onGenerated?.(result.password);
        }
    };

    const input = (
        <input
            type="password"
            value={value}
            onChange={(e) => onChange(e.target.value)}
            placeholder={placeholder}
            required
            autoComplete="new-password"
            aria-invalid={isInvalid}
            aria-describedby={requirementsId}
            className={inputClassName}
            disabled={disabled}
        />
    );

    const generateBtn = (
        <button
            type="button"
            className="password-generate-btn"
            onClick={handleGenerate}
            disabled={disabled || generating}
        >
            {generating ? '…' : 'Сгенерировать'}
        </button>
    );

    const hints = (
        <>
            {generatorUnavailable && (
                <p className="password-hint password-hint--error">Сервис генерации паролей недоступен</p>
            )}
            {generateError && !generatorUnavailable && (
                <p className="password-hint password-hint--error">{generateError}</p>
            )}
            {checking && value && <p className="password-hint">Проверка пароля…</p>}
            {usedFallback && value && !checking && (
                <p className="password-hint password-hint--warn">
                    Проверка локально (сервис валидации недоступен)
                </p>
            )}
            <ul id={requirementsId} className="password-requirements">
                {PASSWORD_REQUIREMENTS.map((req) => (
                    <li
                        key={req.id}
                        className={
                            reqStatus[req.id]
                                ? 'password-requirements__item password-requirements__item--ok'
                                : 'password-requirements__item'
                        }
                    >
                        {req.label}
                    </li>
                ))}
            </ul>
        </>
    );

    if (labelWrapper === 'settings') {
        return (
            <label className="settings-label">
                {label}
                <div className="password-field-row settings-password-row mt-2">
                    {input}
                    {generateBtn}
                </div>
                {hints}
            </label>
        );
    }

    return (
        <label>
            {label}
            <div className="password-field-row">
                {input}
                {generateBtn}
            </div>
            {hints}
        </label>
    );
};

export default NewPasswordField;
