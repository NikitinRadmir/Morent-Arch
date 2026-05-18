import { useState } from 'react';
import { generatePassword } from '../api/passwordApi';
import { useDebouncedPasswordValidation } from './useDebouncedPasswordValidation';
import { generatePasswordFallback, requirementStatus } from '../utils/passwordFallback';

/** Генерация + debounced-валидация нового пароля (как на регистрации). */
export function usePasswordField(password) {
    const [generating, setGenerating] = useState(false);
    const [generatorUnavailable, setGeneratorUnavailable] = useState(false);
    const [generateError, setGenerateError] = useState('');

    const { validation, checking, usedFallback, isValid, isInvalid } = useDebouncedPasswordValidation(password);
    const reqStatus = requirementStatus(validation);

    const generate = async () => {
        setGenerating(true);
        setGeneratorUnavailable(false);
        setGenerateError('');

        const result = await generatePassword();
        setGenerating(false);

        if (!result.ok) {
            if (result.unavailable) {
                const password = generatePasswordFallback();
                setGeneratorUnavailable(true);
                setGenerateError('');
                return { ok: true, password, usedLocalFallback: true };
            }
            setGenerateError(result.error || 'Не удалось сгенерировать пароль');
            return { ok: false, password: null };
        }

        return { ok: true, password: result.password, usedLocalFallback: false };
    };

    const resetGeneratorState = () => {
        setGeneratorUnavailable(false);
        setGenerateError('');
    };

    return {
        validation,
        checking,
        usedFallback,
        isValid,
        isInvalid,
        reqStatus,
        generating,
        generatorUnavailable,
        generateError,
        generate,
        resetGeneratorState,
    };
}
