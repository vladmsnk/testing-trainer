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

func (h *Handler) Register(c *gin.Context) {
	return
}
