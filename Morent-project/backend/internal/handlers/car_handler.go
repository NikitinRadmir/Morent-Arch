package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	carsdto "morent-backend/internal/modules/cars/httpdto"
	"morent-backend/internal/service"
	"gorm.io/gorm"
)

type CarHandler struct {
	service    *service.CarService
	logService *service.LogService
}

func NewCarHandler(service *service.CarService, logService *service.LogService) *CarHandler {
	return &CarHandler{service: service, logService: logService}
}

func (h *CarHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cars, errGetAll := h.service.GetAll()
	if errGetAll != nil {
		http.Error(w, "Error fetching cars: "+errGetAll.Error(), http.StatusInternalServerError)
		return
	}

	jsonData, errEncode := json.Marshal(cars)
	if errEncode != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func (h *CarHandler) GetById(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Path
	parts := strings.Split(path, "/")
	
	if len(parts) < 3 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	idStr := parts[len(parts)-1]
	id, errParse := strconv.ParseUint(idStr, 10, 32)
	if errParse != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	car, errGet := h.service.GetByID(int(id))
	if errGet != nil {
		http.Error(w, "Error fetching car: "+errGet.Error(), http.StatusInternalServerError)
		return
	}
	if car == nil {
		http.Error(w, fmt.Sprintf("Car with ID %d not found", id), http.StatusNotFound)
		return
	}

	jsonData, errEncode := json.Marshal(car)
	if errEncode != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func (h *CarHandler) GetFiltered(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	queryParams := r.URL.Query()
	
	query := carsdto.FilterQuery{}
	
	if nameParam := queryParams.Get("name"); nameParam != "" {
		query.Name = nameParam
	}
	
	if carTypeParam := queryParams.Get("carType"); carTypeParam != "" {
		query.CarType = carTypeParam
	}
	
	if capacityStr := queryParams.Get("capacity"); capacityStr != "" {
		if cap, err := strconv.Atoi(capacityStr); err == nil {
			query.Capacity = &cap
		}
	}
	
	if priceUnderStr := queryParams.Get("priceUnder"); priceUnderStr != "" {
		if price, err := strconv.ParseFloat(priceUnderStr, 64); err == nil {
			query.PriceUnder = &price
		}
	}
	
	filteredCars, errFilter := h.service.GetFiltered(query.Name, query.CarType, query.Capacity, query.PriceUnder)
	if errFilter != nil {
		http.Error(w, "Error fetching filtered cars: "+errFilter.Error(), http.StatusInternalServerError)
		return
	}

	jsonData, errEncode := json.Marshal(filteredCars)
		if errEncode != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func (h *CarHandler) CreateCar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()

	var req carsdto.UpsertCarRequest
	if errDecode := json.NewDecoder(r.Body).Decode(&req); errDecode != nil {
		http.Error(w, "Invalid request body: "+errDecode.Error(), http.StatusBadRequest)
		return
	}
	car := req.Car

	if errCreate := h.service.Create(&car); errCreate != nil {
		_ = h.logService.LogEvent(ctx, service.LogEvent{
			Time:     time.Now(),
			Type:     service.LogCRUD,
			Action:   "create_car",
			ObjectID: car.ID,
			Data:     car,
			Result:   "error",
			Message:  errCreate.Error(),
		})
		http.Error(w, "Error creating car: "+errCreate.Error(), http.StatusInternalServerError)
		return
	}
	_ = h.logService.LogEvent(ctx, service.LogEvent{
		Time:     time.Now(),
		Type:     service.LogCRUD,
		Action:   "create_car",
		ObjectID: car.ID,
		Data:     car,
		Result:   "success",
	})
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(car)
}

func (h *CarHandler) UpdateCar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()

	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	idStr := parts[len(parts)-1]
	id, errParse := strconv.ParseUint(idStr, 10, 32)
	if errParse != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	var req carsdto.UpsertCarRequest
	if errDecode := json.NewDecoder(r.Body).Decode(&req); errDecode != nil {
		http.Error(w, "Invalid request body: "+errDecode.Error(), http.StatusBadRequest)
		return
	}
	car := req.Car
	car.ID = uint(id)

	if errUpdate := h.service.Update(&car); errUpdate != nil {
		_ = h.logService.LogEvent(ctx, service.LogEvent{
			Time:     time.Now(),
			Type:     service.LogCRUD,
			Action:   "update_car",
			ObjectID: car.ID,
			Data:     car,
			Result:   "error",
			Message:  errUpdate.Error(),
		})
		if errors.Is(errUpdate, gorm.ErrRecordNotFound) {
			http.Error(w, "Car not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Error updating car: "+errUpdate.Error(), http.StatusInternalServerError)
		return
	}
	_ = h.logService.LogEvent(ctx, service.LogEvent{
		Time:     time.Now(),
		Type:     service.LogCRUD,
		Action:   "update_car",
		ObjectID: car.ID,
		Data:     car,
		Result:   "success",
	})
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(car)
}

func (h *CarHandler) DeleteCar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()

	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	idStr := parts[len(parts)-1]
	id, errParse := strconv.ParseUint(idStr, 10, 32)
	if errParse != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	if errDelete := h.service.Delete(uint(id)); errDelete != nil {
		_ = h.logService.LogEvent(ctx, service.LogEvent{
			Time:     time.Now(),
			Type:     service.LogCRUD,
			Action:   "delete_car",
			ObjectID: id,
			Result:   "error",
			Message:  errDelete.Error(),
		})
		if errors.Is(errDelete, gorm.ErrRecordNotFound) {
			http.Error(w, "Car not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Error deleting car: "+errDelete.Error(), http.StatusInternalServerError)
		return
	}
	_ = h.logService.LogEvent(ctx, service.LogEvent{
		Time:     time.Now(),
		Type:     service.LogCRUD,
		Action:   "delete_car",
		ObjectID: id,
		Result:   "success",
	})
	w.WriteHeader(http.StatusNoContent)
}

