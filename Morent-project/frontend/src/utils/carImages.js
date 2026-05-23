export const DEFAULT_CAR_IMAGE = '/images/cars/OctaviaRsBlack.png';

const MODEL_IMAGES = [
    ['octavia', '/images/cars/OctaviaRsBlack.png'],
    ['kodiaq', '/images/cars/KodiaqBlue.png'],
    ['kamiq', '/images/cars/KamiqOrange.png'],
    ['kushaq', '/images/cars/KushaqOrange.png'],
    ['rapid', '/images/cars/RapidBlue.png'],
    ['scala', '/images/cars/ScalaBlue.png'],
    ['scout', '/images/cars/ScoutBlue.png'],
    ['fabia', '/images/cars/FabiaRed.png'],
    ['enyaq', '/images/cars/EnyaqGreen.png'],
    ['citigo', '/images/cars/CitigoYellow.png'],
    ['suv', '/images/cars/KodiaqBlue.png'],
];

export const STABLE_UNSPLASH_CAR_IMAGE =
    'https://images.unsplash.com/photo-1492144534655-ae79c964c9d7?auto=format&fit=crop&w=900&q=80';

export const getCarFallbackImage = (...parts) => {
    const text = parts.filter(Boolean).join(' ').toLowerCase();
    const match = MODEL_IMAGES.find(([model]) => text.includes(model));
    return match ? match[1] : DEFAULT_CAR_IMAGE;
};

export const normalizeCarImageSrc = (src, fallback = DEFAULT_CAR_IMAGE, baseUrl = 'http://localhost') => {
    const value = String(src || '').trim();
    if (!value) return fallback;
    if (value === '/images/cars/test.png') return fallback;

    try {
        const url = new URL(value, baseUrl);
        if (url.hostname === 'source.unsplash.com') {
            return STABLE_UNSPLASH_CAR_IMAGE;
        }
    } catch {
        return value;
    }

    return value;
};
