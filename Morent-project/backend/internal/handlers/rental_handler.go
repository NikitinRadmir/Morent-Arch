package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"morent-backend/internal/config"
	"morent-backend/internal/models"
	rentalsdto "morent-backend/internal/modules/rentals/httpdto"
	"morent-backend/internal/modules/transport/http/common"
	"morent-backend/internal/service"
)

type RentalHandler struct {
	authService   *service.AuthService
	rentalService *service.RentalService
	logService    *service.LogService
	cfg           *config.Config
}

var rentalValidator = validator.New()

func NewRentalHandler(authService *service.AuthService, rentalService *service.RentalService, logService *service.LogService, cfg *config.Config) *RentalHandler {
	return &RentalHandler{
		authService:   authService,
		rentalService: rentalService,
		logService:    logService,
		cfg:           cfg,
	}
}

func (h *RentalHandler) List(w http.ResponseWriter, r *http.Request) {
	user, err := h.authenticate(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
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
	user, err := h.authenticate(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	ctx := r.Context()

	var req rentalsdto.RentalRequest
	if errDecode := json.NewDecoder(r.Body).Decode(&req); errDecode != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := rentalValidator.Struct(req); err != nil {
		http.Error(w, "validation error: "+err.Error(), http.StatusBadRequest)
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

	bankToken := common.BankSessionToken(r, h.cfg)
	rental, errCreate := h.rentalService.CreateRental(user.ID, req.CarID, start, end, req.TotalPrice, bankToken)
	if errCreate != nil {
		status, msg := rentalErrorStatus(errCreate)
		_ = h.logService.LogEvent(ctx, service.LogEvent{
			Time:     time.Now(),
			Type:     service.LogRental,
			Action:   "create_rental",
			UserID:   user.ID,
			ObjectID: req.CarID,
			Result:   "error",
			Message:  errCreate.Error(),
		})
		http.Error(w, msg, status)
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
	carID64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
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
	return common.Authenticate(h.authService, h.cfg, r)
}

func rentalErrorStatus(err error) (int, string) {
	switch {
	case errors.Is(err, service.ErrCarNotFound):
		return http.StatusNotFound, "автомобиль не найден"
	case errors.Is(err, service.ErrInvalidRentalPeriod):
		return http.StatusBadRequest, "некорректный период аренды"
	case errors.Is(err, service.ErrRentalPriceMismatch):
		return http.StatusBadRequest, "сумма аренды не совпадает с расчётом на сервере"
	case errors.Is(err, service.ErrCarAlreadyBooked):
		return http.StatusConflict, "автомобиль уже забронирован на выбранные даты"
	case errors.Is(err, service.ErrEmailNotVerified):
		return http.StatusForbidden, "подтвердите email перед бронированием"
	case errors.Is(err, service.ErrBankSessionRequired):
		return http.StatusUnauthorized, "для оплаты откройте Morent Bank и войдите в аккаунт"
	case errors.Is(err, service.ErrBankUnavailable):
		return http.StatusServiceUnavailable, "банковский сервис временно недоступен"
	case errors.Is(err, service.ErrInsufficientBankBalance), errors.Is(err, service.ErrBankInsufficientFunds):
		return http.StatusPaymentRequired, "недостаточно средств на банковском счёте"
	case errors.Is(err, service.ErrBankSessionInvalid):
		return http.StatusUnauthorized, "сессия банка истекла — откройте Morent Bank снова"
	case errors.Is(err, service.ErrBankInvalidAmount):
		return http.StatusBadRequest, "некорректная сумма оплаты"
	default:
		return http.StatusInternalServerError, "не удалось выполнить операцию с арендой"
	}
}
