import React from 'react';

const Options = () => {
  return (
    <div className='row justify-content-center my-5'>

      <div className='col-xl-5 col-10 pr-xl-0'>
        <div className='col-12 mx-auto inner-option p-4'>
          <h3 className='col-12 p-0'><img src="./images/pick-up-icon.png" alt="Pick-Up Icon" /> Pick - Up</h3>
          <div className='row mt-4'>
            <div className='col-4 p-2 gray-border-right'>
              <label htmlFor="locations-pickup">Locations</label>
              <select id="locations-pickup">
                <option value="">Select your city</option>
              </select>
            </div>
            <div className='col-4 p-2 gray-border-right'>
              <label htmlFor="locations-pickup">Date</label>
              <select id="locations-pickup">
                <option value="">Select your date</option>
              </select>
            </div>
            <div className='col-4 p-2'>
              <label htmlFor="locations-pickup">Time</label>
              <select id="locations-pickup">
                <option value="">Select your time</option>
              </select>
            </div>
          </div>
        </div>
      </div>


      <div className='col-xl-1 my-5 my-xl-0 swap-button-div'>
        <button className="swap-button"><img src="/images/swap-icon.png" alt="Swap Icon" /></button>
      </div>


      <div className='col-xl-5 col-10 pl-xl-0'>
        <div className='col-12 mx-auto inner-option p-4'>
        <h3 className='col-12 p-0'><img src="./images/drop-off-icon.png" alt="Drop-Off Icon" /> Drop - Off</h3>
          <div className='row mt-4'>
            <div className='col-4 p-2 gray-border-right'>
              <label htmlFor="locations-pickup">Locations</label>
              <select id="locations-pickup">
                <option value="">Select your city</option>
              </select>
            </div>
            <div className='col-4 p-2 gray-border-right'>
              <label htmlFor="locations-pickup">Date</label>
              <select id="locations-pickup">
                <option value="">Select your date</option>
              </select>
            </div>
            <div className='col-4 p-2'>
              <label htmlFor="locations-pickup">Time</label>
              <select id="locations-pickup">
                <option value="">Select your time</option>
              </select>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Options;
