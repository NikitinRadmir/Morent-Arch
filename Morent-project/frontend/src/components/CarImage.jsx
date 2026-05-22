import React, { useMemo, useState } from 'react';
import { getCarFallbackImage, normalizeCarImageSrc } from '../utils/carImages';

const CarImage = ({ src, alt, className, fallbackKey, onError, ...props }) => {
    const fallback = useMemo(() => getCarFallbackImage(fallbackKey, alt), [fallbackKey, alt]);
    const initialSrc = useMemo(() => normalizeCarImageSrc(src, fallback), [src, fallback]);
    const [currentSrc, setCurrentSrc] = useState(initialSrc);
    const [usedFallback, setUsedFallback] = useState(false);

    React.useEffect(() => {
        setCurrentSrc(initialSrc);
        setUsedFallback(false);
    }, [initialSrc]);

    const handleError = (event) => {
        if (!usedFallback && currentSrc !== fallback) {
            setUsedFallback(true);
            setCurrentSrc(fallback);
            return;
        }
        onError?.(event);
    };

    return (
        <img
            {...props}
            src={currentSrc}
            className={className}
            alt={alt || 'Car'}
            loading={props.loading || 'lazy'}
            onError={handleError}
        />
    );
};

export default CarImage;
