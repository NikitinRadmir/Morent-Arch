export const formatUsd = (value) => {
    if (value == null || Number.isNaN(Number(value))) {
        return '—';
    }
    return new Intl.NumberFormat('en-US', {
        style: 'currency',
        currency: 'USD',
        minimumFractionDigits: 2,
        maximumFractionDigits: 2,
    }).format(Number(value));
};
