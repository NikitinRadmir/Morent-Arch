const RentalSum = ({car, totalAmount}) => {
    return (
        <div className="row p-4">
            <div className="col-12 p-4 form-container">
                <div className="Rental-h">
                    <h5>Rental Summary</h5>
                    <p><highlited-gray>Prices may change depending on the length of the rental and the price of your rental car.</highlited-gray></p>
                </div>
                <div className="row rent-name-img">
                    <div className="rental-img-container mr-4">
                        <img src={car.imgSrc} alt=""/>
                    </div>
                    <div className="rental-name ml-4">
                        <h1 className="pb-3">{car.name}</h1>
                        <div className="rating d-flex align-items-center gap-3 mb-5">
                            <i className="fas fa-star text-warning"></i>
                            <i className="fas fa-star text-warning"></i>
                            <i className="fas fa-star text-warning"></i>
                            <i className="fas fa-star text-warning"></i>
                            <i className="fas fa-star-half-alt text-warning"></i>
                            <span className="text-muted">440+ Reviewer</span>
                        </div>
                    </div>
                </div>
                <div className="row mt-4">
                    <div className="col-6 p-0">
                        <p><highlited-gray>Subtotal</highlited-gray></p>
                    </div>
                    <div className="col-6 t-e p-0">
                        <p>${totalAmount}.00</p>
                    </div>
                    <div className="col-6 p-0">
                        <p><highlited-gray>Tax</highlited-gray></p>
                    </div>
                    <div className="col-6 t-e p-0">
                        <p>$0</p>
                    </div>
                    <div className="rental-apply row">
                        <p className="mb-0"><highlited-gray>Apply promo code</highlited-gray></p>
                        <h5 className="t-e mb-0 ml-auto">Apply now</h5>
                    </div>
                    <div className="col-12 row mt-4">
                        <div>
                            <h5>Total Rental Price</h5>
                            <p><highlited-gray>Overall price and includes rental discount</highlited-gray></p>
                        </div>
                        <div className="ml-auto">
                            <h1 className="mb-0 mt-1">${totalAmount}.00</h1>
                        </div>
                    </div>

                </div>

            </div>
        </div>
    );
};

export default RentalSum;

//AHMEDVARD