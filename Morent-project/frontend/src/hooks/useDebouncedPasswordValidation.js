import { useEffect, useRef, useState } from 'react';
import { validatePassword } from '../api/passwordApi';
import { validatePasswordFallback } from '../utils/passwordFallback';

const DEBOUNCE_MS = 450;

export function useDebouncedPasswordValidation(password) {
    const [validation, setValidation] = useState(null);
    const [checking, setChecking] = useState(false);
    const [usedFallback, setUsedFallback] = useState(false);
    const requestId = useRef(0);

    useEffect(() => {
        if (!password) {
            setValidation(null);
            setChecking(false);
            setUsedFallback(false);
            return undefined;
        }

        const id = ++requestId.current;
        setChecking(true);

        const timer = setTimeout(async () => {
            const remote = await validatePassword(password);
            if (id !== requestId.current) return;

            if (remote.ok && remote.result) {
                setValidation({ ...remote.result, fallback: false });
                setUsedFallback(false);
            } else if (remote.unavailable) {
                setValidation(validatePasswordFallback(password));
                setUsedFallback(true);
            } else {
                setValidation(validatePasswordFallback(password));
                setUsedFallback(true);
            }
            setChecking(false);
        }, DEBOUNCE_MS);

        return () => clearTimeout(timer);
    }, [password]);

    const isValid = validation?.valid === true;
    const isInvalid = password.length > 0 && validation && !validation.valid && !checking;

    return { validation, checking, usedFallback, isValid, isInvalid };
}
