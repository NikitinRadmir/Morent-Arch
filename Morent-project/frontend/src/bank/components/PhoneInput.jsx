import React, { useCallback } from 'react';
import { formatPhoneInputValue } from '../utils/phoneMask';

const PhoneInput = ({ id, name, value, onChange, placeholder = '8 (___) ___-__-__', className = '' }) => {
  const handleChange = useCallback(
    (e) => {
      onChange(formatPhoneInputValue(e.target.value));
    },
    [onChange],
  );

  const handlePaste = useCallback(
    (e) => {
      e.preventDefault();
      const text = e.clipboardData?.getData('text') ?? '';
      onChange(formatPhoneInputValue(text));
    },
    [onChange],
  );

  return (
    <input
      type="tel"
      id={id}
      name={name}
      className={`form-control mb-form-control phone-input ${className}`.trim()}
      placeholder={placeholder}
      value={value}
      onChange={handleChange}
      onPaste={handlePaste}
      inputMode="numeric"
      autoComplete="tel"
    />
  );
};

export default PhoneInput;
