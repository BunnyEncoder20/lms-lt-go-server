package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"go-server/internal/models"
	"go-server/internal/utils"
)

type Handler struct {
	svc Service
	log *slog.Logger
}

func NewHandler(svc Service, logger *slog.Logger) *Handler {
	return &Handler{
		svc: svc,
		log: logger,
	}
}

func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{
			Message: "invalid request body",
		})
		return
	}

	tokenPair, err := h.svc.Login(r.Context(), req.PesNumber, req.Password)
	if err != nil {
		h.log.Debug("invalid pesNumber or password", "err", err)
		utils.WriteJSON(w, http.StatusUnauthorized, models.JSONResponse{
			Message: "invalid email or password",
		})
		return
	}

	isProd := os.Getenv("APP_ENV") == "production"

	http.SetCookie(w, &http.Cookie{
		Name:     "access-token",
		Value:    tokenPair.AccessToken,
		Expires:  time.Now().Add(accessTokenExpiry),
		HttpOnly: true,
		Secure:   isProd,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh-token",
		Value:    tokenPair.RefreshToken,
		Expires:  time.Now().Add(refreshTokenExpiry),
		HttpOnly: true,
		Secure:   isProd,
		SameSite: http.SameSiteStrictMode,
		Path:     "/auth",
	})

	h.log.Info("Login successful", "pesNumber:", req.PesNumber)
	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Message: "login successful",
	})
}

func (h *Handler) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh-token")
	if err != nil {
		utils.WriteJSON(w, http.StatusUnauthorized, models.JSONResponse{
			Success: false,
			Message: "unauthorized: missing refresh token",
		})
		return
	}

	tokenPair, err := h.svc.RefreshToken(r.Context(), cookie.Value)
	if err != nil {
		clearAuthCookies(w)
		utils.WriteJSON(w, http.StatusUnauthorized, models.JSONResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access-token",
		Value:    tokenPair.AccessToken,
		Expires:  time.Now().Add(accessTokenExpiry),
		HttpOnly: true,
		Secure:   os.Getenv("APP_ENV") == "production",
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh-token",
		Value:    tokenPair.RefreshToken,
		Expires:  time.Now().Add(refreshTokenExpiry),
		HttpOnly: true,
		Secure:   os.Getenv("APP_ENV") == "production",
		SameSite: http.SameSiteStrictMode,
		Path:     "/auth",
	})

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Message: "token refreshed successfully",
	})
}

func (h *Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh-token")
	if err != nil {
		h.log.Warn("refresh-token not found")
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{
			Success: false,
			Message: "refresh token not found",
		})
		return
	}

	_ = h.svc.Logout(r.Context(), cookie.Value)

	// clear cookies always (even if DB fails)
	clearAuthCookies(w)

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Message: "logged out successfully",
	})
}

func clearAuthCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access-token",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Path:     "/",
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh-token",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Path:     "/auth",
	})
}
