package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	authdto "morent-backend/internal/modules/auth/httpdto"
	"morent-backend/internal/service"
)

type AuthHandler struct {
	service    *service.AuthService
	logService *service.LogService
}

func NewAuthHandler(service *service.AuthService, logService *service.LogService) *AuthHandler {
	return &AuthHandler{service: service, logService: logService}
}

var authValidator = validator.New()

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req authdto.RegisterRequest
	if errDecode := json.NewDecoder(r.Body).Decode(&req); errDecode != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if err := authValidator.Struct(req); err != nil {
		http.Error(w, "validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	user, token, errRegister := h.service.Register(req.Name, req.Email, req.Password)
	if errRegister != nil {
		_ = h.logService.LogEvent(ctx, service.LogEvent{
			Time:    time.Now(),
			Type:    service.LogAuth,
			Action:  "register",
			Result:  "error",
			Message: errRegister.Error(),
		})
		status := http.StatusInternalServerError
		if errors.Is(errRegister, service.ErrUserExists) {
			status = http.StatusBadRequest
		}
		http.Error(w, errRegister.Error(), status)
		return
	}
	_ = h.logService.LogEvent(ctx, service.LogEvent{
		Time:     time.Now(),
		Type:     service.LogAuth,
		Action:   "register",
		UserID:   user.ID,
		Result:   "success",
		Message:  "user registered",
		ObjectID: nil,
	})

	response := authdto.AuthResponse{
		Token: token,
		User:  *user,
	}
	h.setSessionCookie(w, token)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req authdto.LoginRequest
	if errDecode := json.NewDecoder(r.Body).Decode(&req); errDecode != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if err := authValidator.Struct(req); err != nil {
		http.Error(w, "validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	user, token, errLogin := h.service.Login(req.Email, req.Password)
	if errLogin != nil {
		_ = h.logService.LogEvent(ctx, service.LogEvent{
			Time:    time.Now(),
			Type:    service.LogAuth,
			Action:  "login",
			Result:  "error",
			Message: errLogin.Error(),
		})
		status := http.StatusInternalServerError
		if errors.Is(errLogin, service.ErrInvalidCredentials) {
			status = http.StatusUnauthorized
		}
		http.Error(w, errLogin.Error(), status)
		return
	}
	_ = h.logService.LogEvent(ctx, service.LogEvent{
		Time:    time.Now(),
		Type:    service.LogAuth,
		Action:  "login",
		UserID:  user.ID,
		Result:  "success",
		Message: "user logged in",
	})

	response := authdto.AuthResponse{
		Token: token,
		User:  *user,
	}
	h.setSessionCookie(w, token)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	token := h.extractToken(r)
	if token == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}
	lower := strings.ToLower(token)
	if strings.HasPrefix(lower, "bearer ") {
		token = strings.TrimSpace(token[7:])
	}

	if err := h.service.Logout(token); err != nil {
		_ = h.logService.LogEvent(ctx, service.LogEvent{
			Time:    time.Now(),
			Type:    service.LogAuth,
			Action:  "logout",
			Result:  "error",
			Message: err.Error(),
		})
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	_ = h.logService.LogEvent(ctx, service.LogEvent{
		Time:    time.Now(),
		Type:    service.LogAuth,
		Action:  "logout",
		Result:  "success",
		Message: "user logged out",
	})
	h.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Profile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	token := h.extractToken(r)
	if token == "" {
		_ = h.logService.LogEvent(ctx, service.LogEvent{
			Time:    time.Now(),
			Type:    service.LogAuth,
			Action:  "profile",
			Result:  "error",
			Message: "missing token",
		})
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}
	lower := strings.ToLower(token)
	if strings.HasPrefix(lower, "bearer ") {
		token = strings.TrimSpace(token[7:])
	}
	user, err := h.service.GetUserByToken(token)
	if err != nil || user == nil {
		_ = h.logService.LogEvent(ctx, service.LogEvent{
			Time:    time.Now(),
			Type:    service.LogAuth,
			Action:  "profile",
			Result:  "error",
			Message: "invalid token",
		})
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}
	_ = h.logService.LogEvent(ctx, service.LogEvent{
		Time:    time.Now(),
		Type:    service.LogAuth,
		Action:  "profile",
		UserID:  user.ID,
		Result:  "success",
		Message: "profile accessed",
	})
	resp := user.ToResponse()
	json.NewEncoder(w).Encode(resp)
}

// UpdateProfile updates basic user profile fields.
func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	token := h.extractToken(r)
	if token == "" {
		_ = h.logService.LogEvent(ctx, service.LogEvent{
			Time:    time.Now(),
			Type:    service.LogProfile,
			Action:  "update_profile",
			Result:  "error",
			Message: "missing token",
		})
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}
	lower := strings.ToLower(token)
	if strings.HasPrefix(lower, "bearer ") {
		token = strings.TrimSpace(token[7:])
	}
	user, err := h.service.GetUserByToken(token)
	if err != nil || user == nil {
		_ = h.logService.LogEvent(ctx, service.LogEvent{
			Time:    time.Now(),
			Type:    service.LogProfile,
			Action:  "update_profile",
			Result:  "error",
			Message: "invalid token",
		})
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	var req authdto.ProfileUpdateRequest
	if errDecode := json.NewDecoder(r.Body).Decode(&req); errDecode != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updated, errUpdate := h.service.UpdateProfile(user.ID, req.Name, req.Nickname, req.Position, req.AvatarURL)
	if errUpdate != nil {
		_ = h.logService.LogEvent(ctx, service.LogEvent{
			Time:    time.Now(),
			Type:    service.LogProfile,
			Action:  "update_profile",
			UserID:  user.ID,
			Result:  "error",
			Message: errUpdate.Error(),
		})
		http.Error(w, errUpdate.Error(), http.StatusInternalServerError)
		return
	}
	_ = h.logService.LogEvent(ctx, service.LogEvent{
		Time:   time.Now(),
		Type:   service.LogProfile,
		Action: "update_profile",
		UserID: user.ID,
		Result: "success",
		Data:   updated,
	})
	json.NewEncoder(w).Encode(updated)
}

// ChangePassword changes user's password.
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	token := h.extractToken(r)
	if token == "" {
		_ = h.logService.LogEvent(ctx, service.LogEvent{
			Time:    time.Now(),
			Type:    service.LogProfile,
			Action:  "change_password",
			Result:  "error",
			Message: "missing token",
		})
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}
	lower := strings.ToLower(token)
	if strings.HasPrefix(lower, "bearer ") {
		token = strings.TrimSpace(token[7:])
	}
	user, err := h.service.GetUserByToken(token)
	if err != nil || user == nil {
		_ = h.logService.LogEvent(ctx, service.LogEvent{
			Time:    time.Now(),
			Type:    service.LogProfile,
			Action:  "change_password",
			Result:  "error",
			Message: "invalid token",
		})
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	var req authdto.PasswordChangeRequest
	if errDecode := json.NewDecoder(r.Body).Decode(&req); errDecode != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if errChange := h.service.ChangePassword(user.ID, req.OldPassword, req.NewPassword); errChange != nil {
		_ = h.logService.LogEvent(ctx, service.LogEvent{
			Time:    time.Now(),
			Type:    service.LogProfile,
			Action:  "change_password",
			UserID:  user.ID,
			Result:  "error",
			Message: errChange.Error(),
		})
		status := http.StatusInternalServerError
		if errors.Is(errChange, service.ErrInvalidCredentials) {
			status = http.StatusUnauthorized
		}
		http.Error(w, errChange.Error(), status)
		return
	}
	_ = h.logService.LogEvent(ctx, service.LogEvent{
		Time:   time.Now(),
		Type:   service.LogProfile,
		Action: "change_password",
		UserID: user.ID,
		Result: "success",
	})
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) setSessionCookie(w http.ResponseWriter, token string) {
	if token == "" {
		return
	}
	cookieName := strings.TrimSpace(os.Getenv("SESSION_COOKIE_NAME"))
	if cookieName == "" {
		cookieName = "morent_session"
	}
	secure := strings.EqualFold(strings.TrimSpace(os.Getenv("SESSION_COOKIE_SECURE")), "true")
	domain := strings.TrimSpace(os.Getenv("SESSION_COOKIE_DOMAIN"))
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
		Domain:   domain,
	})
}

func (h *AuthHandler) clearSessionCookie(w http.ResponseWriter) {
	cookieName := strings.TrimSpace(os.Getenv("SESSION_COOKIE_NAME"))
	if cookieName == "" {
		cookieName = "morent_session"
	}
	domain := strings.TrimSpace(os.Getenv("SESSION_COOKIE_DOMAIN"))
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Domain:   domain,
		MaxAge:   -1,
	})
}

func (h *AuthHandler) extractToken(r *http.Request) string {
	token := strings.TrimSpace(r.Header.Get("Authorization"))
	if token == "" {
		cookieName := strings.TrimSpace(os.Getenv("SESSION_COOKIE_NAME"))
		if cookieName == "" {
			cookieName = "morent_session"
		}
		cookie, err := r.Cookie(cookieName)
		if err == nil && cookie != nil {
			token = strings.TrimSpace(cookie.Value)
		}
	}
	lower := strings.ToLower(token)
	if strings.HasPrefix(lower, "bearer ") {
		token = strings.TrimSpace(token[7:])
	}
	return token
}
