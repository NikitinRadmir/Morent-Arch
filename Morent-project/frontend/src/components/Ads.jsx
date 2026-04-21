import React from 'react';

const Ads = () => {
    return (
      <div className="additional row justify-content-center">
        <div className='col-xl-5 col-lg-6 col-9 mr-xl-5 pl-xl-0'>
          <div className="ads_1">
            <div className="ads">
              <div className="title_ads">
                <h1>The Best Platform for Car Rental</h1>
                <p>Ease of doing a car rental safely and reliably. Of course at a low price.</p>
                <button className="rent">Rental Car</button>
              </div>
            </div>
          </div>
        </div>
        
        <div className='col-xl-5 col-lg-6 col-9 ml-xl-5 pr-xl-0'>
          <div className="ads_2">
            <div className="ads_for_2">
              <div className="title_ads_2">
                <h1>Easy way to rent a car at a low price</h1>
                <p>Providing cheap car rental services and safe and comfortable facilities.</p>
                <button className="rent_2">Rental Car</button>
              </div>
            </div>
          </div>
        </div>
        
      </div>
    );
}

export default Ads;