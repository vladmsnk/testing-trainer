package habit

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"testing_trainer/internal/entities"
	"testing_trainer/utils/token"
)

type Handler struct {
	uc UseCase
}

type UseCase interface {
	CreateHabit(ctx context.Context, username string, habit entities.Habit) (int64, error)
	ListUserHabits(ctx context.Context, username string) ([]entities.Habit, error)
}

func NewHabitHandler(r *gin.RouterGroup, uc UseCase) {
	h := Handler{uc: uc}

	r.POST("/habits", h.CreateHabit)
	r.GET("/habits", h.ListUserHabits)
}

// CreateHabit godoc
// @Summary create habit endpoint
// @Schemes
// @Description Creates habit in the system
// @Tags example
// @Accept json
// @Produce json
// @Param requestBody body CreateHabitRequest true "Create habit"
// @Param Authorization header string true "Bearer"
// @Success 200 {string} ok
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /tracker/habits [post]
func (h *Handler) CreateHabit(c *gin.Context) {
	var createHabitRequest CreateHabitRequest

	username, err := token.ExtractUsernameFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
	}

	if err := c.ShouldBindJSON(&createHabitRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := createHabitRequest.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	id, err := h.uc.CreateHabit(c, username, toEntityHabit(createHabitRequest))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusCreated, gin.H{"id": strconv.Itoa(int(id))})
}

// ListUserHabits godoc
// @Summary list user habits endpoint
// @Schemes
// @Description Lists all habits for the authenticated user
// @Tags example
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer"
// @Success 200 {array} ListUserHabitsResponse "List of user habits"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Router /tracker/habits [get]
func (h *Handler) ListUserHabits(c *gin.Context) {
	username, err := token.ExtractUsernameFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
	}

	habits, err := h.uc.ListUserHabits(c, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, toListUserHabitsResponse(username, habits))
}
