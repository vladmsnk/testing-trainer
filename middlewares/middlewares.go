package middlewares

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"testing_trainer/internal/entities"
	"testing_trainer/internal/usecase/user"
	"testing_trainer/utils/token"
)

type AuthUseCase interface {
	GetToken(ctx context.Context, username string) (entities.Token, error)
}

func AuthMiddleware(authUc user.UseCase) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := token.TokenValidInContext(c)
		if err != nil {
			c.String(http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}

		tokenUserName, err := token.ExtractUsernameFromToken(c)
		if err != nil {
			c.String(http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}

		tokenFromHeaders := token.ExtractToken(c)

		tokenEntityFromStorage, err := authUc.GetToken(c, tokenUserName)
		if err != nil {
			c.String(http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}

		if tokenEntityFromStorage.AccessToken != tokenFromHeaders {
			c.String(http.StatusUnauthorized, "Invalid token")
			c.Abort()
			return
		}

		c.Set("username", tokenUserName)
		c.Next()
	}
}
