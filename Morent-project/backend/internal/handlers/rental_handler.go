package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"morent-backend/internal/models"
	rentalsdto "morent-backend/internal/modules/rentals/httpdto"
	"morent-backend/internal/service"
)

type RentalHandler struct {
	authService   *service.AuthService
	rentalService *service.RentalService
	logService    *service.LogService
}

var rentalValidator = validator.New()

func NewRentalHandler(authService *service.AuthService, rentalService *service.RentalService, logService *service.LogService) *RentalHandler {
	return &RentalHandler{
		authService:   authService,
		rentalService: rentalService,
		logService:    logService,
	}
}

func (h *RentalHandler) List(w http.ResponseWriter, r *http.Request) {
	user, authErr := h.authenticate(r)
	if authErr != nil {
		http.Error(w, authErr.Error(), http.StatusUnauthorized)
		return
	}
	ctx := r.Context()

	rentals, errList := h.rentalService.ListRentals(user.ID)
	if errList != nil {
		_ = h.logService.LogEvent(ctx, service.LogEvent{
			Time:    time.Now(),
			Type:    service.LogRental,
			Action:  "list_rentals",
			UserID:  user.ID,
			Result:  "error",
			Message: errList.Error(),
		})
		http.Error(w, "failed to load rentals: "+errList.Error(), http.StatusInternalServerError)
		return
	}
	_ = h.logService.LogEvent(ctx, service.LogEvent{
		Time:   time.Now(),
		Type:   service.LogRental,
		Action: "list_rentals",
		UserID: user.ID,
		Result: "success",
	})
	json.NewEncoder(w).Encode(rentals)
}

func (h *RentalHandler) Create(w http.ResponseWriter, r *http.Request) {
	user, authErr := h.authenticate(r)
	if authErr != nil {
		http.Error(w, authErr.Error(), http.StatusUnauthorized)
		return
	}
	ctx := r.Context()

	var req rentalsdto.RentalRequest
	if errDecode := json.NewDecoder(r.Body).Decode(&req); errDecode != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if validationErr := rentalValidator.Struct(req); validationErr != nil {
		http.Error(w, "validation error: "+validationErr.Error(), http.StatusBadRequest)
		return
	}

	start, errStart := time.Parse("2006-01-02", req.StartDate)
	if errStart != nil {
		http.Error(w, "invalid startDate format", http.StatusBadRequest)
		return
	}
	end, errEnd := time.Parse("2006-01-02", req.EndDate)
	if errEnd != nil {
		http.Error(w, "invalid endDate format", http.StatusBadRequest)
		return
	}

	rental, errCreate := h.rentalService.CreateRental(user.ID, req.CarID, start, end, req.TotalPrice)
	if errCreate != nil {
		status := http.StatusInternalServerError
		if errCreate == service.ErrCarNotFound || errCreate == service.ErrInvalidRentalPeriod || errCreate == service.ErrCarAlreadyBooked {
			status = http.StatusBadRequest
		}
		_ = h.logService.LogEvent(ctx, service.LogEvent{
			Time:     time.Now(),
			Type:     service.LogRental,
			Action:   "create_rental",
			UserID:   user.ID,
			ObjectID: req.CarID,
			Result:   "error",
			Message:  errCreate.Error(),
		})
		http.Error(w, errCreate.Error(), status)
		return
	}
	_ = h.logService.LogEvent(ctx, service.LogEvent{
		Time:     time.Now(),
		Type:     service.LogRental,
		Action:   "create_rental",
		UserID:   user.ID,
		ObjectID: rental.ID,
		Result:   "success",
		Data:     rental,
	})
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rental)
}

// BookedDates returns booked periods for a specific car (public, no auth).
// URL pattern: /Rentals/Car/{id}
func (h *RentalHandler) BookedDates(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	idStr := parts[len(parts)-1]
	carID64, carIDParseErr := strconv.ParseUint(idStr, 10, 64)
	if carIDParseErr != nil {
		http.Error(w, "invalid car id", http.StatusBadRequest)
		return
	}

	bookings, errList := h.rentalService.ListCarBookings(uint(carID64))
	if errList != nil {
		http.Error(w, "failed to load bookings: "+errList.Error(), http.StatusInternalServerError)
		return
	}

	resp := make([]rentalsdto.BookingDTO, 0, len(bookings))
	for _, b := range bookings {
		resp = append(resp, rentalsdto.BookingDTO{
			StartDate: b.StartDate,
			EndDate:   b.EndDate,
		})
	}

	json.NewEncoder(w).Encode(resp)
}

func (h *RentalHandler) authenticate(r *http.Request) (*models.User, error) {
	token := strings.TrimSpace(r.Header.Get("Authorization"))
	if token == "" {
		return nil, service.ErrInvalidToken
	}
	lower := strings.ToLower(token)
	if strings.HasPrefix(lower, "bearer ") {
		token = strings.TrimSpace(token[7:])
	}
	user, getUserByTokenErr := h.authService.GetUserByToken(token)
	if getUserByTokenErr != nil {
		return nil, getUserByTokenErr
	}
	return user, nil
}
