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
}

func NewAuthHandler(r *gin.RouterGroup, uc UseCase) {
	h := Handler{
		uc: uc,
	}

	r.POST("/register", h.Register)
	r.POST("/login", h.Login)
}

// Register godoc
// @Summary register endpoint
// @Schemes
// @Description Registers users in the system
// @Tags example
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
// @Tags example
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

func (h *Handler) Logout(c *gin.Context) {
}
