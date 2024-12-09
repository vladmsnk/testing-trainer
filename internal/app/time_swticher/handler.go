package time_swticher

import (
	"net/http"
	"testing_trainer/utils/token"

	"github.com/gin-gonic/gin"
	"testing_trainer/internal/usecase/time_switcher"
)

type Handler struct {
	timeSwitcherUseCase time_switcher.UseCase
}

func NewHandler(r *gin.RouterGroup, tsu time_switcher.UseCase) *Handler {
	h := &Handler{
		timeSwitcherUseCase: tsu,
	}

	r.POST("/next-day", h.SwitchToNextDay)
	r.GET("/current-time", h.GetCurrentTime)
	r.PUT("/reset-time", h.ResetToCurrentDay)
	return h
}

// ResetToCurrentDay godoc
// @Summary reset time to current day endpoint
// @Schemes
// @Description Resets the time to the current day
// @Tags time
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer"
// @Success 200 {string} ok
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /time/reset-time [put]
func (h *Handler) ResetToCurrentDay(c *gin.Context) {
	username, err := token.ExtractUsernameFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	err = h.timeSwitcherUseCase.ResetToCurrentDay(c, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	c.JSON(200, gin.H{"message": "success"})
}

// GetCurrentTime godoc
// @Summary get current time endpoint
// @Schemes
// @Description Returns the current time
// @Tags time
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer"
// @Success 200 {string} ok
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /time/current-time [get]
func (h *Handler) GetCurrentTime(c *gin.Context) {
	username, err := token.ExtractUsernameFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	currentTime, err := h.timeSwitcherUseCase.GetCurrentTime(c, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	c.JSON(200, gin.H{"currentTime": currentTime})

}

// SwitchToNextDay godoc
// @Summary switch to next day endpoint
// @Schemes
// @Description Switches to the next day
// @Tags time
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer"
// @Success 200 {string} ok
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /time/next-day [post]
func (h *Handler) SwitchToNextDay(c *gin.Context) {
	username, err := token.ExtractUsernameFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	err = h.timeSwitcherUseCase.SwitchToNextDay(c, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	c.JSON(200, gin.H{"message": "success"})
}
