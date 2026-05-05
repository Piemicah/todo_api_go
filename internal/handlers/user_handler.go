package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"
	"todo_api/internal/config"
	"todo_api/internal/models"
	"todo_api/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// ---------------------------------------------------------------------------
// Request types
// ---------------------------------------------------------------------------

type RegisterRequest struct {
	Email    string      `json:"email"    binding:"required"`
	Password string      `json:"password" binding:"required"`
	Role     models.Role `json:"role"`
}

type LoginRequest struct {
	Email    string `json:"email"    binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ---------------------------------------------------------------------------
// Cookie config
// ---------------------------------------------------------------------------

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 7 * 24 * time.Hour

	accessCookieName  = "access_token"
	refreshCookieName = "refresh_token"
)

// setAuthCookies writes both tokens as HttpOnly, Secure, SameSite=Strict cookies.
func setAuthCookies(ctx *gin.Context, cfg *config.Config, accessToken, refreshToken string) {
	secure := cfg.Env == "production" // only send over HTTPS in prod

	ctx.SetSameSite(http.SameSiteStrictMode)

	ctx.SetCookie(
		accessCookieName,
		accessToken,
		int(accessTokenTTL.Seconds()), // MaxAge in seconds
		"/",                           // path — sent on every request
		cfg.Domain,                    // e.g. "example.com" or "" for localhost
		secure,                        // Secure flag
		true,                          // HttpOnly — JS cannot read this
	)

	ctx.SetCookie(
		refreshCookieName,
		refreshToken,
		int(refreshTokenTTL.Seconds()),
		"/auth/refresh", // scoped to refresh endpoint only — not sent on every request
		cfg.Domain,
		secure,
		true,
	)
}

// clearAuthCookies expires both cookies immediately.
func clearAuthCookies(ctx *gin.Context, cfg *config.Config) {
	secure := cfg.Env == "production"
	ctx.SetSameSite(http.SameSiteStrictMode)
	ctx.SetCookie(accessCookieName, "", -1, "/", cfg.Domain, secure, true)
	ctx.SetCookie(refreshCookieName, "", -1, "/auth/refresh", cfg.Domain, secure, true)
}

// ---------------------------------------------------------------------------
// Token helpers
// ---------------------------------------------------------------------------

func generateRefreshToken() (string, error) {
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func buildAccessToken(user *models.User, secret string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    string(user.Role),
		"exp":     time.Now().Add(accessTokenTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func issueAndSetCookies(pool *pgxpool.Pool, user *models.User, cfg *config.Config, ctx *gin.Context) error {
	accessToken, err := buildAccessToken(user, cfg.JWTSecret)
	if err != nil {
		return err
	}

	rawRefresh, err := generateRefreshToken()
	if err != nil {
		return err
	}

	rt := &models.RefreshToken{
		UserID:    user.ID,
		Token:     rawRefresh,
		ExpiresAt: time.Now().Add(refreshTokenTTL),
	}
	if _, err = repository.CreateRefreshToken(pool, rt); err != nil {
		return err
	}

	setAuthCookies(ctx, cfg, accessToken, rawRefresh)
	return nil
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

// CreateUserHandler  POST /auth/register
func CreateUserHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req RegisterRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if len(req.Password) < 6 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 6 characters"})
			return
		}
		if req.Role == models.RoleAdmin {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "cannot self-assign admin role"})
			return
		}

		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}

		user := &models.User{Email: req.Email, Password: string(hashed), Role: req.Role}
		created, err := repository.CreateUser(pool, user)
		if err != nil {
			if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
				ctx.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusCreated, created)
	}
}

func GetAllUsersHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		users, err := repository.GetAllUsers(pool)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, users)
	}
}

// LoginHandler  POST /auth/login
// Sets HttpOnly cookies — returns no tokens in the body.
func LoginHandler(pool *pgxpool.Pool, cfg *config.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req LoginRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user, err := repository.GetUserByEmail(pool, req.Email)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		if err = issueAndSetCookies(pool, user, cfg, ctx); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
			return
		}

		// Return only non-sensitive user info — never the tokens.
		ctx.JSON(http.StatusOK, gin.H{
			"id":    user.ID,
			"email": user.Email,
			"role":  user.Role,
		})
	}
}

// RefreshHandler  POST /auth/refresh
// Reads the refresh token from its cookie, rotates the pair, sets new cookies.
func RefreshHandler(pool *pgxpool.Pool, cfg *config.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		rawToken, err := ctx.Cookie(refreshCookieName)
		if err != nil || rawToken == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token cookie missing"})
			return
		}

		rt, err := repository.GetRefreshToken(pool, rawToken)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
			return
		}
		if rt.Revoked {
			_ = repository.RevokeAllUserTokens(pool, rt.UserID)
			clearAuthCookies(ctx, cfg)
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token already used"})
			return
		}
		if time.Now().After(rt.ExpiresAt) {
			clearAuthCookies(ctx, cfg)
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token expired"})
			return
		}

		if err = repository.RevokeRefreshToken(pool, rt.Token); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not revoke old token"})
			return
		}

		user, err := repository.GetUserByID(pool, rt.UserID)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}

		if err = issueAndSetCookies(pool, user, cfg, ctx); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"message": "tokens refreshed"})
	}
}

// LogoutHandler  POST /auth/logout
// Revokes the DB token and clears both cookies.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
	AllDevices   bool   `json:"all_devices"`
}

func LogoutHandler(pool *pgxpool.Pool, cfg *config.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var body LogoutRequest
		_ = ctx.ShouldBindJSON(&body)

		// 1. Try cookie first (browser clients)
		rawToken, _ := ctx.Cookie(refreshCookieName)

		// 2. Fall back to request body (REST clients, mobile, etc.)
		if rawToken == "" {
			rawToken = body.RefreshToken
		}

		if rawToken != "" {
			if body.AllDevices {
				if userIDVal, exists := ctx.Get("user_id"); exists {
					if userID, ok := userIDVal.(string); ok {
						_ = repository.RevokeAllUserTokens(pool, userID)
					}
				}
			} else {
				_ = repository.RevokeRefreshToken(pool, rawToken)
			}
		}

		clearAuthCookies(ctx, cfg)
		ctx.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
	}
}

// TestProtectedHandler  GET /protected-test
func TestProtectedHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID, _ := ctx.Get("user_id")
		role, _ := ctx.Get("role")
		ctx.JSON(http.StatusOK, gin.H{
			"message": "protected route accessed successfully",
			"user_id": userID,
			"role":    role,
		})
	}
}