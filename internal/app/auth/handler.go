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
}

func NewAuthHandler(r *gin.RouterGroup, uc UseCase) {
	h := Handler{
		uc: uc,
	}

	r.POST("/register", h.Register)
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

	return
}
