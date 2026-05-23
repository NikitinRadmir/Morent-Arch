import assert from 'node:assert/strict';
import test from 'node:test';

import { messageForStatus, normalizeErrorMessage, resilientFetch, resilientJson } from './apiClient.js';

test('messageForStatus maps service failures to friendly messages', () => {
    assert.match(messageForStatus(500), /Внутренняя ошибка/);
    assert.match(messageForStatus(502), /прокси/);
    assert.match(messageForStatus(503), /временно недоступен/);
    assert.match(messageForStatus(504), /ожидания/);
    assert.equal(messageForStatus(404), null);
});

test('resilientJson returns null for successful empty body', async () => {
    const originalFetch = globalThis.fetch;
    globalThis.fetch = async () => new Response('', { status: 200 });
    try {
        assert.equal(await resilientJson('/empty'), null);
    } finally {
        globalThis.fetch = originalFetch;
    }
});

test('resilientFetch throws safe message for 503 responses', async () => {
    const originalFetch = globalThis.fetch;
    globalThis.fetch = async () => new Response('', { status: 503, statusText: 'Service Unavailable' });
    try {
        await assert.rejects(
            () => resilientFetch('/down'),
            (error) => {
                assert.equal(error.status, 503);
                assert.equal(error.code, 'service_unavailable');
                assert.match(error.message, /временно недоступен/);
                return true;
            },
        );
    } finally {
        globalThis.fetch = originalFetch;
    }
});

test('normalizeErrorMessage hides raw validator output', () => {
    const raw = "validation error: Key: 'LoginRequest.Password' Error:Field validation for 'Password' failed on the 'min' tag";
    assert.equal(normalizeErrorMessage(raw, 400), 'Проверьте корректность заполнения полей');
});
