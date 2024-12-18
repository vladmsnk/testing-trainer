package progress

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"testing_trainer/internal/entities"
	"testing_trainer/internal/usecase/progress"
	"testing_trainer/utils/token"
)

type Handler struct {
	uc UseCase
}

type UseCase interface {
	GetHabitProgress(ctx context.Context, username string, habitId int) (entities.ProgressWithGoal, error)
	AddHabitProgress(ctx context.Context, username string, habitId int) error
	GetCurrentProgressForAllUserHabits(ctx context.Context, username string) ([]entities.CurrentPeriodProgress, error)
}

func NewProgressHandler(r *gin.RouterGroup, uc UseCase) {
	h := Handler{uc: uc}

	r.POST("/progress/:habitId", h.AddProgress)
	r.GET("/progress/:habitId", h.GetHabitProgress)
	r.GET("/reminder", h.GetReminder)
}

// AddProgress godoc
// @Summary add progress endpoint
// @Schemes
// @Description Adds progress to the habit
// @Tags progress
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer"
// @Param habitId path string true "Habit ID"
// @Success 200 {string} ok
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /tracker/progress/{habitId} [post]
func (h *Handler) AddProgress(c *gin.Context) {
	username, err := token.ExtractUsernameFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	habitIdStr := c.Param("habitId")
	if habitIdStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "habitId is required"})
		return
	}

	habitId, err := strconv.Atoi(habitIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "habitId must be an integer"})
		return
	}

	err = h.uc.AddHabitProgress(c, username, habitId)
	if err != nil {
		if errors.Is(err, progress.ErrGoalCompleted) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, progress.ErrHabitNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An error occurred while adding habit progress"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

// GetHabitProgress godoc
// @Summary get progress endpoint
// @Schemes
// @Description Get progress of the habit
// @Tags progress
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer"
// @Param habitId path string true "Habit ID"
// @Success 200 {string} ok
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /tracker/progress/{habitId} [get]
func (h *Handler) GetHabitProgress(c *gin.Context) {
	username, err := token.ExtractUsernameFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	habitIdStr := c.Param("habitId")
	if habitIdStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "habitID is required"})
		return
	}

	habitId, err := strconv.Atoi(habitIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "habitID must be an integer"})
		return
	}

	progressWithGoal, err := h.uc.GetHabitProgress(c, username, habitId)
	if err != nil {
		if errors.Is(err, progress.ErrHabitNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, progress.ErrGoalNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, toHabitProgressResponse(progressWithGoal))
}

// GetReminder godoc
// @Summary get reminder endpoint
// @Schemes
// @Description Get reminder for the user
// @Tags progress
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer"
// @Success 200 {string} ok
// @Failure 500 {string} string "Internal Server Error"
// @Router /tracker/reminder [get]
func (h *Handler) GetReminder(c *gin.Context) {
	username, err := token.ExtractUsernameFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	currentProgressForAllUserHabits, err := h.uc.GetCurrentProgressForAllUserHabits(c, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An error occurred while retrieving habit progress"})
		return
	}

	c.JSON(http.StatusOK, toReminderResponse(currentProgressForAllUserHabits))
}
