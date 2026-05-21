export function normalizeCardDigits(raw) {
  return String(raw || '').replace(/\D/g, '').slice(0, 16);
}

export function formatCardInputValue(raw) {
  const digits = normalizeCardDigits(raw);
  const parts = [];
  for (let i = 0; i < digits.length; i += 4) {
    parts.push(digits.slice(i, i + 4));
  }
  return parts.join(' ');
}
