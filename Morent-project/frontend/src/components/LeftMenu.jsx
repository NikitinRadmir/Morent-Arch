import React, { useState, useEffect } from 'react';

const LeftMenu = ({ props, onFilterChange, initialFilters }) => {
    // Инициализация состояния фильтров из initialFilters
    const [filters, setFilters] = useState({
        types: initialFilters.types || [],
        capacities: initialFilters.capacities || [],
        maxPrice: initialFilters.maxPrice || 200,
    });

    // Обновление состояния фильтров при изменении initialFilters
    useEffect(() => {
        setFilters({
            types: initialFilters.types || [],
            capacities: initialFilters.capacities || [],
            maxPrice: initialFilters.maxPrice || 200,
        });
    }, [initialFilters]);

    const typeCounts = props.reduce((counts, prop) => {
        counts[prop.type] = (counts[prop.type] || 0) + 1;
        return counts;
    }, {});

    const capacityCounts = props.reduce((counts, prop) => {
        counts[prop.capacity] = (counts[prop.capacity] || 0) + 1;
        return counts;
    }, {});

    const handleTypeChange = (type) => {
        setFilters((prevFilters) => {
            const updatedTypes = prevFilters.types.includes(type)
                ? prevFilters.types.filter((t) => t !== type)
                : [...prevFilters.types, type];
            const newFilters = { ...prevFilters, types: updatedTypes };
            onFilterChange(newFilters);
            return newFilters;
        });
    };

    const handleCapacityChange = (capacity) => {
        setFilters((prevFilters) => {
            const updatedCapacities = prevFilters.capacities.includes(capacity)
                ? prevFilters.capacities.filter((c) => c !== capacity)
                : [...prevFilters.capacities, capacity];
            const newFilters = { ...prevFilters, capacities: updatedCapacities };
            onFilterChange(newFilters);
            return newFilters;
        });
    };

    const handlePriceChange = (event) => {
        const newPrice = event.target.value;
        setFilters((prevFilters) => {
            const newFilters = { ...prevFilters, maxPrice: parseInt(newPrice, 10) };
            onFilterChange(newFilters);
            return newFilters;
        });
    };

    return (
        <div className="left-menu">
            <div className="section">
                <h3>TYPE</h3>
                {Object.keys(typeCounts).map((type) => (
                    <label key={type} className="checkbox-label">
                        <input
                            type="checkbox"
                            checked={filters.types.includes(type)}
                            onChange={() => handleTypeChange(type)}
                        />
                        {type} ({typeCounts[type]})
                    </label>
                ))}
            </div>

            <div className="section">
                <h3>CAPACITY</h3>
                {Object.keys(capacityCounts).map((capacity) => (
                    <label key={capacity} className="checkbox-label">
                        <input
                            type="checkbox"
                            checked={filters.capacities.includes(capacity)}
                            onChange={() => handleCapacityChange(capacity)}
                        />
                        {capacity} Person ({capacityCounts[capacity]})
                    </label>
                ))}
            </div>

            <div className="section">
                <h3>PRICE</h3>
                <input
                    type="range"
                    min="0"
                    max="200"
                    value={filters.maxPrice}
                    onChange={handlePriceChange}
                    className="price-slider"
                />
                <p>Max. ${filters.maxPrice}.00</p>
            </div>
        </div>
    );
};

export default LeftMenu;