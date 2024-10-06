package habit

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"testing_trainer/internal/entities"
	"testing_trainer/utils/token"
)

type Handler struct {
	uc UseCase
}

type UseCase interface {
	CreateHabit(ctx context.Context, username string, habit entities.Habit) (int64, error)
}

func NewHabitHandler(r *gin.RouterGroup, uc UseCase) {
	h := Handler{uc: uc}

	r.POST("/habits", h.CreateHabit)
}

// CreateHabit godoc
// @Summary create habit endpoint
// @Schemes
// @Description Creates habit in the system
// @Tags example
// @Accept json
// @Produce json
// @Param requestBody body CreateHabitRequest true "Create habit"
// @Success 200 {string} ok
// @Router /tracker/habits [post]
func (h *Handler) CreateHabit(c *gin.Context) {
	var createHabit entities.Habit

	username, err := token.ExtractUsernameFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
	}

	if err := c.ShouldBindJSON(&createHabit); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := createHabit.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	_, err = h.uc.CreateHabit(c, username, createHabit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Habit created"})
}
