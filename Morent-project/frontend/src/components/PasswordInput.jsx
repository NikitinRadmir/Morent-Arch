import React, { useState } from 'react';

const EyeOpenIcon = () => (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" aria-hidden="true">
        <path
            d="M2 12C2 12 5.5 5 12 5C18.5 5 22 12 22 12C22 12 18.5 19 12 19C5.5 19 2 12 2 12Z"
            stroke="currentColor"
            strokeWidth="1.75"
            strokeLinecap="round"
            strokeLinejoin="round"
        />
        <circle cx="12" cy="12" r="3.25" stroke="currentColor" strokeWidth="1.75" />
    </svg>
);

const EyeClosedIcon = () => (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" aria-hidden="true">
        <path
            d="M3 3L21 21M10.58 10.58C10.21 10.95 10 11.45 10 12C10 13.1 10.9 14 12 14C12.55 14 13.05 13.79 13.42 13.42M6.71 6.71C4.66 8.17 3.29 10.23 2 12C2 12 5.5 19 12 19C14.05 19 15.92 18.36 17.49 17.35M9.88 5.09C10.57 5.03 11.28 5 12 5C18.5 5 22 12 22 12C21.27 13.39 20.17 14.62 18.87 15.63"
            stroke="currentColor"
            strokeWidth="1.75"
            strokeLinecap="round"
            strokeLinejoin="round"
        />
    </svg>
);

/** Поле пароля с кнопкой «показать / скрыть» внутри инпута. */
const PasswordInput = ({
    value,
    onChange,
    name,
    id,
    placeholder,
    disabled = false,
    required = false,
    autoComplete = 'new-password',
    'aria-invalid': ariaInvalid,
    'aria-describedby': ariaDescribedby,
    className = '',
    inputClassName = '',
}) => {
    const [visible, setVisible] = useState(false);
    const wrapClass = ['password-input-wrap', className].filter(Boolean).join(' ');
    const fieldClass = ['password-input-wrap__field', inputClassName].filter(Boolean).join(' ');

    return (
        <div className={wrapClass}>
            <input
                type={visible ? 'text' : 'password'}
                name={name}
                id={id}
                value={value}
                onChange={onChange}
                placeholder={placeholder}
                disabled={disabled}
                required={required}
                autoComplete={autoComplete}
                aria-invalid={ariaInvalid}
                aria-describedby={ariaDescribedby}
                className={fieldClass}
            />
            <button
                type="button"
                className="password-visibility-btn"
                onClick={() => setVisible((v) => !v)}
                disabled={disabled}
                aria-label={visible ? 'Скрыть пароль' : 'Показать пароль'}
                aria-pressed={visible}
            >
                {visible ? <EyeClosedIcon /> : <EyeOpenIcon />}
            </button>
        </div>
    );
};

export default PasswordInput;
