import test from 'node:test';
import assert from 'node:assert/strict';

import {
    DEFAULT_CAR_IMAGE,
    STABLE_UNSPLASH_CAR_IMAGE,
    getCarFallbackImage,
    normalizeCarImageSrc,
} from './carImages.js';

test('getCarFallbackImage chooses a matching local model image', () => {
    assert.equal(getCarFallbackImage('Skoda Kodiaq 2024'), '/images/cars/KodiaqBlue.png');
    assert.equal(getCarFallbackImage('Fabia hatchback'), '/images/cars/FabiaRed.png');
    assert.equal(getCarFallbackImage('Payment Test SUV'), '/images/cars/KodiaqBlue.png');
});

test('getCarFallbackImage falls back to a default local car image', () => {
    assert.equal(getCarFallbackImage('Unknown model'), DEFAULT_CAR_IMAGE);
});

test('normalizeCarImageSrc replaces source.unsplash.com with stable images.unsplash.com URL', () => {
    assert.equal(
        normalizeCarImageSrc('https://source.unsplash.com/900x500/?car', DEFAULT_CAR_IMAGE),
        STABLE_UNSPLASH_CAR_IMAGE,
    );
});

test('normalizeCarImageSrc preserves valid external and local URLs', () => {
    assert.equal(
        normalizeCarImageSrc('https://images.unsplash.com/photo-1', DEFAULT_CAR_IMAGE),
        'https://images.unsplash.com/photo-1',
    );
    assert.equal(
        normalizeCarImageSrc('https://www.regcheck.org.uk/image.aspx/@2016_toyota_camry', DEFAULT_CAR_IMAGE),
        'https://www.regcheck.org.uk/image.aspx/@2016_toyota_camry',
    );
    assert.equal(normalizeCarImageSrc('/images/cars/RapidBlue.png'), '/images/cars/RapidBlue.png');
});

test('normalizeCarImageSrc uses fallback for empty values', () => {
    assert.equal(normalizeCarImageSrc('', '/fallback.png'), '/fallback.png');
    assert.equal(normalizeCarImageSrc(null, '/fallback.png'), '/fallback.png');
    assert.equal(normalizeCarImageSrc('/images/cars/test.png', '/fallback.png'), '/fallback.png');
});
