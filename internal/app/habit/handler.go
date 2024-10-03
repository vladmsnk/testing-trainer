package habit

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
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

func (h *Handler) CreateHabit(c *gin.Context) {
	var habit entities.Habit

	username, err := token.ExtractUsernameFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
	}

	if err := c.ShouldBindJSON(&habit); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := habit.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	_, err = h.uc.CreateHabit(c, username, habit)
	if err != nil {
		//todo handle error
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Habit created"})
}
