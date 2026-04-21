import React, { useContext, useState } from 'react';
import toast from 'react-hot-toast';
import { AuthContext } from '../context/AuthContext';

const Reviews = ({ reviews, carId, onCommentAdded }) => {
    const [visibleReviews, setVisibleReviews] = useState(1);
    const [newComment, setNewComment] = useState('');
    const [rating, setRating] = useState(5);
    const [submitting, setSubmitting] = useState(false);
    const [error, setError] = useState('');
    const { isAuthenticated, authRequest } = useContext(AuthContext);

    const handleShowAll = () => {
        setVisibleReviews(reviews.length);
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        setError('');
        if (!newComment.trim()) return;
            try {
                setSubmitting(true);
                const created = await authRequest('/comments', {
                method: 'POST',
                body: JSON.stringify({ carId, description: newComment, rating }),
            });
            setNewComment('');
            setRating(5);
            if (onCommentAdded) {
                onCommentAdded(created);
            }
            toast.success('Комментарий сохранён');
        } catch (err) {
            const msg = err.message || 'Failed to add comment';
            setError(msg);
            toast.error(`Ошибка добавления комментария: ${msg}`);
        } finally {
            setSubmitting(false);
        }
    };

    return (
        <div className='col-12 p-3'>
            <div className='p-3 reviews'>
                <h4 className='m-3 mb-4'>
                    Reviews <span className='background-blue'>{reviews.length}</span>
                </h4>

                {isAuthenticated && (
                    <form className="mb-4" onSubmit={handleSubmit}>
                        <textarea
                            className="form-container-input"
                            style={{ height: '80px', resize: 'vertical' }}
                            placeholder="Share your experience with this car..."
                            value={newComment}
                            onChange={(e) => setNewComment(e.target.value)}
                        />
                        <div className="mt-2 d-flex align-items-center">
                            <span className="mr-2">Your rating:</span>
                            <select
                                className="form-container-input"
                                style={{ width: '120px', height: '40px' }}
                                value={rating}
                                onChange={(e) => setRating(Number(e.target.value))}
                            >
                                {[5, 4, 3, 2, 1].map((val) => (
                                    <option key={val} value={val}>
                                        {val} / 5
                                    </option>
                                ))}
                            </select>
                        </div>
                        {error && <div className="auth-alert auth-alert--error mt-2">{error}</div>}
                        <button
                            type="submit"
                            className="auth-submit mt-2"
                            disabled={submitting}
                        >
                            {submitting ? 'Sending...' : 'Add comment'}
                        </button>
                    </form>
                )}

                {reviews.slice(0, visibleReviews).map((review) => (
                    <div key={review.id} className='review row'>
                        <div className='review-img col-1 p-0'>
                            <img src={review.photo} className="reviewer-img" alt="reviewer" />
                        </div>
                        <div className='col-11 row p-0'>
                            <div className='col-10 p-0'>
                                <h5 className='reviewer-name'>{review.name}</h5>
                                {review.post && <h6 className='reviewer-post'>{review.post}</h6>}
                            </div>
                            <div className='col-2 p-0 data-stars'>
                                <p>{review.date}</p>
                                <div>
                                    {[1,2,3,4,5].map((i) => (
                                        <i
                                            key={i}
                                            className={
                                                review.rating && review.rating >= i
                                                    ? 'fa-solid fa-star'
                                                    : 'fa-regular fa-star'
                                            }
                                        ></i>
                                    ))}
                                </div>
                            </div>
                            <p className='mt-3'>{review.description}</p>
                        </div>
                    </div>
                ))}

                {visibleReviews < reviews.length && (
                    <div className='col-2 offset-5 d-flex align-items-center justify-content-center'>
                        <button
                            className='btn btn-primary'
                            onClick={handleShowAll}
                        >
                            Show All <i className="fa-solid fa-arrow-down ml-2"></i>
                        </button>
                    </div>
                )}
            </div>
        </div>
    );
};

export default Reviews;
