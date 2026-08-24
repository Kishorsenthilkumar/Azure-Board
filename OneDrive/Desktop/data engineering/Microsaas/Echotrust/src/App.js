import React from 'react';
import Navbar from './components/Navbar';
import HeroSection from './components/HeroSection';
import ProductOverview from './components/ProductOverview';
import Footer from './components/Footer';
import Dashboard from './components/Dashboard';
import './styles.css';

const App = () => {
    return (
        <div className="App">
            <Navbar />
            <HeroSection />
            <ProductOverview />
            <Footer />
        </div>
    );
};

export default App; 