import React from 'react';
import { formatCardInputValue } from '../utils/cardMask';

const CardInput = ({ id, name, value, onChange, placeholder = '0000 0000 0000 0000', className = '' }) => {
  const handleChange = (e) => {
    onChange(formatCardInputValue(e.target.value));
  };

  const handlePaste = (e) => {
    e.preventDefault();
    const text = e.clipboardData.getData('text');
    onChange(formatCardInputValue(text));
  };

  return (
    <input
      type="text"
      className={`form-control mb-form-control card-input ${className}`.trim()}
      id={id}
      name={name}
      value={value}
      onChange={handleChange}
      onPaste={handlePaste}
      placeholder={placeholder}
      inputMode="numeric"
      autoComplete="off"
      maxLength={19}
    />
  );
};

export default CardInput;
