// Package handler provides HTTP handlers for the application.
package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/ttani03/gotha-boilerplate/internal/config"
	"github.com/ttani03/gotha-boilerplate/internal/db/generated"
	"github.com/ttani03/gotha-boilerplate/web/templates/pages"
)

// AuthHandler handles authentication-related requests.
type AuthHandler struct {
	db      *pgxpool.Pool
	queries *generated.Queries
	cfg     *config.Config
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(db *pgxpool.Pool, queries *generated.Queries, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		db:      db,
		queries: queries,
		cfg:     cfg,
	}
}

// RegisterPage renders the registration page.
func (h *AuthHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	render(w, r, pages.Register(""))
}

// Register handles user registration.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		render(w, r, pages.Register("フォームの解析に失敗しました"))
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")
	name := r.FormValue("name")

	if email == "" || password == "" || name == "" {
		render(w, r, pages.Register("すべてのフィールドを入力してください"))
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		render(w, r, pages.Register("内部エラーが発生しました"))
		return
	}

	user, err := h.queries.CreateUser(r.Context(), generated.CreateUserParams{
		Email:        email,
		PasswordHash: string(hashedPassword),
		Name:         name,
	})
	if err != nil {
		render(w, r, pages.Register("このメールアドレスは既に登録されています"))
		return
	}

	if err := h.setTokenCookies(w, r, uuidToString(user.ID)); err != nil {
		render(w, r, pages.Register("内部エラーが発生しました"))
		return
	}

	http.Redirect(w, r, "/todos", http.StatusSeeOther)
}

// LoginPage renders the login page.
func (h *AuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	render(w, r, pages.Login(""))
}

// Login handles user login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		render(w, r, pages.Login("フォームの解析に失敗しました"))
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	user, err := h.queries.GetUserByEmail(r.Context(), email)
	if err != nil {
		render(w, r, pages.Login("メールアドレスまたはパスワードが正しくありません"))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		render(w, r, pages.Login("メールアドレスまたはパスワードが正しくありません"))
		return
	}

	if err := h.setTokenCookies(w, r, uuidToString(user.ID)); err != nil {
		render(w, r, pages.Login("内部エラーが発生しました"))
		return
	}

	http.Redirect(w, r, "/todos", http.StatusSeeOther)
}

// Logout handles user logout.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Invalidate refresh token
	if cookie, err := r.Cookie("refresh_token"); err == nil {
		hash := hashToken(cookie.Value)
		_ = h.queries.DeleteRefreshTokenByHash(r.Context(), hash)
	}

	// Clear cookies
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// RefreshToken handles token refresh.
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	hash := hashToken(cookie.Value)
	token, err := h.queries.GetRefreshTokenByHash(r.Context(), hash)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Delete old refresh token
	_ = h.queries.DeleteRefreshTokenByHash(r.Context(), hash)

	// Issue new tokens
	if err := h.setTokenCookies(w, r, uuidToString(token.UserID)); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/todos", http.StatusSeeOther)
}

// setTokenCookies generates and sets access/refresh token cookies.
func (h *AuthHandler) setTokenCookies(w http.ResponseWriter, r *http.Request, userID string) error {
	// Generate access token
	accessToken, err := h.generateAccessToken(userID)
	if err != nil {
		return err
	}

	// Generate refresh token
	refreshToken, err := h.generateRefreshToken(r, userID)
	if err != nil {
		return err
	}

	isSecure := h.cfg.Env == "production"

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		MaxAge:   int(h.cfg.JWTAccessTokenDuration.Seconds()),
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		MaxAge:   int(h.cfg.JWTRefreshTokenDuration.Seconds()),
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteLaxMode,
	})

	return nil
}

// generateAccessToken creates a new JWT access token.
func (h *AuthHandler) generateAccessToken(userID string) (string, error) {
	claims := &jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.cfg.JWTAccessTokenDuration)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.cfg.JWTSecret))
}

// generateRefreshToken creates a new refresh token and stores it in the database.
func (h *AuthHandler) generateRefreshToken(r *http.Request, userID string) (string, error) {
	// Generate random token
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	rawToken := hex.EncodeToString(b)

	// Store hashed token in database
	uid, err := parseUUID(userID)
	if err != nil {
		return "", err
	}

	hash := hashToken(rawToken)
	expiresAt := pgtype.Timestamptz{
		Time:  time.Now().Add(h.cfg.JWTRefreshTokenDuration),
		Valid: true,
	}
	_, err = h.queries.CreateRefreshToken(r.Context(), generated.CreateRefreshTokenParams{
		UserID:    uid,
		TokenHash: hash,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return "", err
	}

	return rawToken, nil
}

// hashToken creates a SHA-256 hash of the token.
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// parseUUID parses a string UUID into a pgtype.UUID.
func parseUUID(s string) (pgtype.UUID, error) {
	var uid pgtype.UUID
	if err := uid.Scan(s); err != nil {
		return uid, err
	}
	return uid, nil
}

// uuidToString converts a pgtype.UUID to a string.
func uuidToString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	// Format as standard UUID string
	b := u.Bytes
	return hex.EncodeToString(b[0:4]) + "-" +
		hex.EncodeToString(b[4:6]) + "-" +
		hex.EncodeToString(b[6:8]) + "-" +
		hex.EncodeToString(b[8:10]) + "-" +
		hex.EncodeToString(b[10:16])
}

// render is a helper to render a templ component.
func render(w http.ResponseWriter, r *http.Request, component templ.Component) {
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
