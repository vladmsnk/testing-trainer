package auth

import (
	"github.com/gin-gonic/gin"
)

type Handler struct {
	uc UseCase
}

type UseCase interface {
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
// @Description register users in the system
// @Tags example
// @Accept json
// @Produce json
// @Success 200 {string} ok
// @Router /register [post]
func (h *Handler) Register(c *gin.Context) {
	return
}
