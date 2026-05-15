const MAX_DIGITS = 11;

export function normalizeDigits(raw) {
  let d = String(raw).replace(/\D/g, '');
  if (d.length === 0) return '';
  if (d[0] === '7') d = `8${d.slice(1)}`;
  else if (d[0] !== '8') d = `8${d}`;
  return d.slice(0, MAX_DIGITS);
}

export function formatPhone(digits) {
  if (!digits || digits.length === 0) return '';
  const v = digits;
  let formatted = v[0];
  if (v.length > 1) {
    formatted += ` (${v.slice(1, Math.min(4, v.length))}`;
  }
  if (v.length >= 4) {
    formatted += ')';
  }
  if (v.length > 4) {
    formatted += ` ${v.slice(4, Math.min(7, v.length))}`;
  }
  if (v.length > 7) {
    formatted += `-${v.slice(7, Math.min(9, v.length))}`;
  }
  if (v.length > 9) {
    formatted += `-${v.slice(9, Math.min(11, v.length))}`;
  }
  return formatted;
}

export function formatPhoneInputValue(raw) {
  return formatPhone(normalizeDigits(raw));
}
