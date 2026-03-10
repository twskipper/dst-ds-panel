package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"dst-ds-panel/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

type AuthHandler struct {
	auth config.Auth
}

func NewAuthHandler(auth config.Auth) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expiresAt"`
}

func (a *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if req.Username != a.auth.Username || req.Password != a.auth.Password {
		writeError(w, http.StatusUnauthorized, "invalid username or password")
		return
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	claims := jwt.RegisteredClaims{
		Subject:   req.Username,
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(a.auth.Secret))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	writeJSON(w, http.StatusOK, LoginResponse{
		Token:     tokenStr,
		ExpiresAt: expiresAt.Unix(),
	})
}

func (a *AuthHandler) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := ""

		// Check Authorization header
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		}

		// Fallback: check query param (for WebSocket connections)
		if tokenStr == "" {
			tokenStr = r.URL.Query().Get("token")
		}

		if tokenStr == "" {
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return []byte(a.auth.Secret), nil
		})
		if err != nil || !token.Valid {
			writeError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		next.ServeHTTP(w, r)
	})
}
