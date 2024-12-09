package auth

import (
	"context"
	"errors"
	"net/http"
	"testing_trainer/internal/usecase/user"

	"github.com/gin-gonic/gin"
	"testing_trainer/internal/entities"
)

type Handler struct {
	uc UseCase
}

type UseCase interface {
	RegisterUser(ctx context.Context, user entities.RegisterUser) error
	Login(ctx context.Context, user entities.User) (entities.Token, error)
	Logout(ctx context.Context, token entities.Token) error
	RefreshToken(ctx context.Context, tkn entities.Token) (entities.Token, error)
}

func NewAuthHandler(r *gin.RouterGroup, uc UseCase) {
	h := Handler{
		uc: uc,
	}

	r.POST("/register", h.Register)
	r.POST("/login", h.Login)
	r.POST("/logout", h.Logout)
	r.POST("/refresh", h.RefreshToken)
}

// Register godoc
// @Summary register endpoint
// @Schemes
// @Description Registers users in the system
// @Tags auth
// @Accept json
// @Produce json
// @Param requestBody body RegisterRequest true "Register user"
// @Success 200 {string} ok
// @Failure 409 {string} string "User already exists"
// @Router /auth/register [post]
func (h *Handler) Register(c *gin.Context) {

	var req RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	registerUser := toEntityRegisterUser(req)
	err := h.uc.RegisterUser(c, registerUser)
	if err != nil {
		if errors.Is(err, user.ErrUserAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User registered"})
	return
}

// Login godoc
// @Summary login endpoint
// @Schemes
// @Description Authenticates users and provides a token
// @Tags auth
// @Accept json
// @Produce json
// @Param requestBody body LoginRequest true "Login user"
// @Success 200 {string} token "JWT Token"
// @Failure 401 {string} string "Unauthorized"
// @Router /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.uc.Login(c, toEntityUser(req))
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, user.ErrInvalidPassword) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": toLoginResponse(token)})
}

// Logout godoc
// @Summary logout endpoint
// @Schemes
// @Description Logs out users from the system
// @Tags auth
// @Accept json
// @Produce json
// @Param requestBody body LogoutRequest true "Logout"
// @Success 200 {string} ok
// @Failure 401 {string} string "Unauthorized"
// @Router /auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	var req LogoutRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.uc.Logout(c, entities.Token{AccessToken: req.AccessToken})
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) || errors.Is(err, user.ErrTokenNotFound) || errors.Is(err, user.ErrInvalidToken) {
			c.JSON(http.StatusOK, gin.H{"message": "User logged out"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User logged out"})
	return
}

// RefreshToken godoc
// @Summary refresh token endpoint
// @Schemes
// @Description Refreshes the authentication token
// @Tags auth
// @Accept json
// @Produce json
// @Param requestBody body RefreshTokenRequest true "Refresh token"
// @Success 200 {string} token "New JWT Token"
// @Failure 401 {string} string "Unauthorized"
// @Router /auth/refresh [post]
func (h *Handler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.uc.RefreshToken(c, entities.Token{RefreshToken: req.RefreshToken})
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": toRefreshResponse(token)})
}
