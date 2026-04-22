package handlers

import (
	"log/slog"
	"net/http"

	"docker-visual/internal/auth"
	"docker-visual/internal/models"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	store  *auth.UserStore
	jwt    *auth.JWTService
	logger *slog.Logger
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(store *auth.UserStore, jwt *auth.JWTService) *AuthHandler {
	return &AuthHandler{
		store:  store,
		jwt:    jwt,
		logger: slog.Default(),
	}
}

// Register creates a new user account.
func (h *AuthHandler) Register(c *gin.Context) {
	var req auth.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error(), Code: "INVALID_INPUT"})
		return
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to process password", Code: "INTERNAL_ERROR"})
		return
	}

	// Create user
	user, err := h.store.CreateUser(req.Username, hashedPassword)
	if err == auth.ErrUserExists {
		c.JSON(http.StatusConflict, models.ErrorResponse{Error: "username already taken", Code: "USER_EXISTS"})
		return
	}
	if err != nil {
		h.logger.Error("failed to create user", "error", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to create user", Code: "INTERNAL_ERROR"})
		return
	}

	// Generate token
	token, expiresAt, err := h.jwt.GenerateToken(user)
	if err != nil {
		h.logger.Error("failed to generate token", "error", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to generate token", Code: "INTERNAL_ERROR"})
		return
	}

	h.logger.Info("user registered", "username", user.Username)
	c.JSON(http.StatusCreated, auth.AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      *user,
	})
}

// Login authenticates a user and returns a JWT token.
func (h *AuthHandler) Login(c *gin.Context) {
	var req auth.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error(), Code: "INVALID_INPUT"})
		return
	}

	// Get user
	user, err := h.store.GetUserByUsername(req.Username)
	if err == auth.ErrUserNotFound {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "invalid credentials", Code: "INVALID_CREDENTIALS"})
		return
	}
	if err != nil {
		h.logger.Error("failed to get user", "error", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "authentication failed", Code: "INTERNAL_ERROR"})
		return
	}

	// Check password
	if !auth.CheckPassword(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "invalid credentials", Code: "INVALID_CREDENTIALS"})
		return
	}

	// Generate token
	token, expiresAt, err := h.jwt.GenerateToken(user)
	if err != nil {
		h.logger.Error("failed to generate token", "error", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to generate token", Code: "INTERNAL_ERROR"})
		return
	}

	// Clear password before returning
	user.Password = ""

	h.logger.Info("user logged in", "username", user.Username)
	c.JSON(http.StatusOK, auth.AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      *user,
	})
}

// Me returns the current authenticated user.
func (h *AuthHandler) Me(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "not authenticated", Code: "UNAUTHORIZED"})
		return
	}

	user, err := h.store.GetUserByID(userID.(int64))
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "user not found", Code: "NOT_FOUND"})
		return
	}

	user.Password = ""
	c.JSON(http.StatusOK, user)
}

// RequiresSetup returns whether initial setup is needed (no users exist).
func (h *AuthHandler) RequiresSetup(c *gin.Context) {
	count, err := h.store.UserCount()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to check setup status", Code: "INTERNAL_ERROR"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"requires_setup": count == 0})
}
