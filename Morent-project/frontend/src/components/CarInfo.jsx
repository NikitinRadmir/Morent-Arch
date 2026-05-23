import React, { useState } from "react";
import { Link } from "react-router-dom";
import CarImage from "./CarImage";
import { getCarFallbackImage, normalizeCarImageSrc } from "../utils/carImages";
import { formatUsd } from "../utils/formatMoney";

const CarInfo = ({ props }) => {
  const fallbackImage = getCarFallbackImage(props.name, props.type);
  const carImage = normalizeCarImageSrc(props.imgSrc, fallbackImage);
  const [selectedImage, setSelectedImage] = useState(null);

  const openModal = (imageSrc) => {
    setSelectedImage(imageSrc);
  };

  const closeModal = () => {
    setSelectedImage(null);
  };

  return (
    <div className="container-fluid mb-5">
      <div className="row align-items-center">
        <div className="col-md-6">
          <div
            className="banner-content"
            style={{
              backgroundColor: "#3563E9",
              height: "360px",
              margin: "40px 0 0 0",
              borderRadius: "10px",
            }}
          >
            <div className="p-4 text-white">
              <h2 className="fw-bold">
                Skoda car with the best <br />
                design and acceleration
              </h2>
              <p>
                Safety and comfort while driving a <br />
                futuristic and elegant sports car
              </p>
              <CarImage
                src={props.imgSrc}
                alt="Current Car"
                className="carImg"
                fallbackKey={`${props.name} ${props.type}`}
              />
            </div>
          </div>

          <div className="thumbnails mt-3 row">
            <div className="col-4 pl-0">
              <CarImage
                src={props.imgSrc}
                alt={props.name}
                className="rounded border border-primary img-thumbnail"
                onClick={() => openModal(carImage)}
                fallbackKey={`${props.name} ${props.type}`}
                style={{ cursor: "pointer" }}
              />
            </div>
            <div className="col-4 px-0">
              <img
                src="/images/salon1.png"
                alt={props.name}
                className="rounded border border-primary img-thumbnail"
                onClick={() => openModal("/images/salon1.png")}
                style={{ cursor: "pointer" }}
              />
            </div>
            <div className="col-4 pr-0">
              <img
                src="/images/salon2.png"
                alt={props.name}
                className="rounded border border-primary img-thumbnail"
                onClick={() => openModal("/images/salon2.png")}
                style={{ cursor: "pointer" }}
              />
            </div>
          </div>
        </div>

        <div className="col-md-6">
          <div className="card rounded p-4">
            <h1 className="card-title fw-bold">{props.name}</h1>
            <div className="rating d-flex align-items-center gap-3 mb-5">
              <i className="fas fa-star text-warning"></i>
              <i className="fas fa-star text-warning"></i>
              <i className="fas fa-star text-warning"></i>
              <i className="fas fa-star text-warning"></i>
              <i className="fas fa-star-half-alt text-warning"></i>
              <span className="text-muted">440+ Reviewer</span>
            </div>
            <p className="card-text text-muted">{props.description}</p>

            <div className="row mt-4">
              <div className="col-3">
                <p style={{ color: "#90A3BF" }}>Type Car</p>
                <p style={{ color: "#90A3BF" }}>Steering</p>
              </div>
              <div className="col-3">
                <p>{props.type}</p>
                <p>{props.transmission}</p>
              </div>
              <div className="col-3">
                <p style={{ color: "#90A3BF" }}>Capacity</p>
                <p style={{ color: "#90A3BF" }}>Gasoline</p>
              </div>
              <div className="col-3">
                <p>{props.capacity} People</p>
                <p>{props.fuel}L</p>
              </div>
            </div>

            <div className="d-flex justify-content-between align-items-center mt-5">
              <h2 className="fw-bold price">
                {formatUsd(props.price)} /<span className="highlited-gray">day</span>
              </h2>

              <a href="#">
                <Link className="btn btn-primary btn-lg" to={`/rent/${props.id}`}>
                  Rent now
                </Link>
              </a>
            </div>
          </div>
        </div>
      </div>

      {selectedImage && (
        <div
          className="modal fade show"
          style={{ display: "block", backgroundColor: "rgba(0, 0, 0, 0.5)" }}
        >
          <div className="modal-dialog modal-dialog-centered modal-lg">
            <div className="modal-content">
              <div className="modal-header">
                <h5 className="modal-title">Image Preview</h5>
                <button
                  type="button"
                  className="btn-close"
                  onClick={closeModal}
                ></button>
              </div>
              <div className="modal-body text-center">
                <CarImage
                  src={selectedImage}
                  alt="Preview"
                  className="img-fluid"
                  fallbackKey={`${props.name} ${props.type}`}
                />
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default CarInfo;
