import React from 'react';
import './ProductOverview.css';

const ProductOverview = () => {
    return (
        <div className="product-overview">
            <h2>Our Products</h2>
            <div className="product-card">
                <h3>Reviews AI</h3>
                <p>Generate, monitor, manage, and share reviews on 200+ sites.</p>
            </div>
            <div className="product-card">
                <h3>Social AI</h3>
                <p>Create, schedule, and monitor social posts across locations.</p>
            </div>
            <div className="product-card">
                <h3>Listings AI</h3>
                <p>Get discovered in local search and outrank the competition.</p>
            </div>
            {/* Add more product cards as needed */}
        </div>
    );
};

export default ProductOverview; 