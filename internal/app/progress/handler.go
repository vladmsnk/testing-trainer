package progress

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
	GetHabitProgress(ctx context.Context, username, habitName string) (entities.ProgressWithGoal, error)
	AddHabitProgress(ctx context.Context, username, habitName string) error
}

func NewProgressHandler(r *gin.RouterGroup, uc UseCase) {
	h := Handler{uc: uc}

	r.POST("/progress/:habitName", h.AddProgress)
	r.GET("/progress/:habitName", h.GetHabitProgress)
}

// AddProgress godoc
// @Summary add progress endpoint
// @Schemes
// @Description Adds progress to the habit
// @Tags example
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer"
// @Param habitName path string true "Habit name"
// @Success 200 {string} ok
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /tracker/progress/{habitName} [post]
func (h *Handler) AddProgress(c *gin.Context) {
	username, err := token.ExtractUsernameFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
	}

	habitName := c.Param("habitName")
	if habitName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "habitName is required"})
		return
	}

	err = h.uc.AddHabitProgress(c, username, habitName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}

// GetHabitProgress godoc
// @Summary get progress endpoint
// @Schemes
// @Description Get progress of the habit
// @Tags example
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer"
// @Param habitName path string true "Habit name"
// @Success 200 {string} ok
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /tracker/progress/{habitName} [get]
func (h *Handler) GetHabitProgress(c *gin.Context) {
	username, err := token.ExtractUsernameFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
	}

	habitName := c.Param("habitName")
	if habitName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "habitName is required"})
		return
	}

	progressWithGoal, err := h.uc.GetHabitProgress(c, username, habitName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toHabitProgressResponse(progressWithGoal))
}
