import React from 'react';
import './Navbar.css'; // Create a separate CSS file for styles

const Navbar = () => {
    return (
        <nav className="navbar">
            <div className="logo">YourLogo</div>
            <ul className="nav-links">
                <li>Solutions</li>
                <li>Products</li>
                <li>Pricing</li>
                <li>Resources</li>
                <li>Contact</li>
            </ul>
        </nav>
    );
};

export default Navbar; 