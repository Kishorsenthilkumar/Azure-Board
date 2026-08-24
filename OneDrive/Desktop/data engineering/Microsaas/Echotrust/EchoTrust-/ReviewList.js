import React from 'react';
import './styles.css'; // Import your styles

// Sample review data
const reviews = [
    { id: 1, text: "Great product!", isFake: false },
    { id: 2, text: "Terrible service, do not buy!", isFake: true, reason: "Suspicious language" },
    // ... other reviews
];

const ReviewList = () => {
    return (
        <div>
            <h2>Customer Reviews</h2>
            <ul>
                {reviews.map(review => (
                    <li key={review.id} style={{ marginBottom: '10px' }}>
                        <p>{review.text}</p>
                        {review.isFake && (
                            <span className="fake-review-warning">
                                ⚠️ Fake Review Detected: {review.reason}
                            </span>
                        )}
                    </li>
                ))}
            </ul>
        </div>
    );
};

export default ReviewList; 