package token

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
)

func TestTokenValid(t *testing.T) {
	var (
		username   = "username"
		testSecret = "testSecret"
	)

	os.Setenv(keyEnvApiSecret, testSecret)
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)

	t.Run("valid token", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"authorized": true,
			"username":   username,
			"exp":        time.Now().Add(time.Hour * 1).Unix(),
		})

		tokenString, err := token.SignedString([]byte("testSecret"))
		assert.NoError(t, err, "Error signing token")

		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		c.Request = req

		err = TokenValid(c)
		assert.NoError(t, err, "Valid token should pass validation")
	})

	t.Run("wrong secret", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"authorized": true,
			"username":   username,
			"exp":        time.Now().Add(time.Hour * 1).Unix(),
		})

		tokenString, err := token.SignedString([]byte("wrongSecret"))
		assert.NoError(t, err, "Error signing token")

		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		c.Request = req

		err = TokenValid(c)
		assert.Error(t, err, "Invalid secret should fail validation")
	})

	t.Run("expired token", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"authorized": true,
			"username":   username,
			"exp":        time.Now().Add(time.Hour * -1).Unix(),
		})

		tokenString, err := token.SignedString([]byte("testSecret"))
		assert.NoError(t, err, "Error signing token")

		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		c.Request = req

		err = TokenValid(c)
		assert.Error(t, err, "Expired token should fail validation")
	})
}

func TestGenerateTokenByUsername(t *testing.T) {
	var (
		//username   = "username"
		testSecret = "testSecret"
	)

	os.Setenv(keyEnvApiSecret, testSecret)
	os.Setenv(keyEnvTokenLifeSpan, "1")
	gin.SetMode(gin.TestMode)

	t.Run("successful generation", func(t *testing.T) {
		//tokenString, err := GenerateTokenByUsername(username)
		//assert.NoError(t, err, "Token generation should not fail")
		//assert.NotEmpty(t, tokenString, "Token string should not be empty")
		//
		//token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		//	return []byte(testSecret), nil
		//})
		//assert.NoError(t, err, "Error should not occur during token parsing")
		//assert.True(t, token.Valid, "Token should be valid")
		//
		//claims, ok := token.Claims.(jwt.MapClaims)
		//assert.True(t, ok, "Claims should be of type jwt.MapClaims")
		//assert.True(t, claims["authorized"].(bool), "Authorized claim should be true")
	})
}

func TestExtractUsernameFromToken(t *testing.T) {
	os.Setenv(keyEnvApiSecret, "testsecret")

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)

	t.Run("ValidToken", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"authorized": true,
			"username":   "testuser",
			"exp":        time.Now().Add(time.Hour * 1).Unix(), // expires in 1 hour
		})

		tokenString, err := token.SignedString([]byte("testsecret"))
		assert.NoError(t, err, "Error should not occur when signing the token")

		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		c.Request = req

		username, err := ExtractUsernameFromToken(c)
		assert.NoError(t, err, "Valid token should pass without error")
		assert.Equal(t, "testuser", username, "Username should be 'testuser'")
	})

	t.Run("MissingUsernameClaim", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"authorized": true,
			"exp":        time.Now().Add(time.Hour * 1).Unix(),
		})

		tokenString, err := token.SignedString([]byte("testsecret"))
		assert.NoError(t, err, "Error should not occur when signing the token")

		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		c.Request = req

		username, err := ExtractUsernameFromToken(c)
		assert.Error(t, err, "Token without username claim should fail")
		assert.Empty(t, username, "Username should be empty when claim is missing")
		assert.Contains(t, err.Error(), "username claim is missing or not a string", "Error message should indicate missing username claim")
	})
}
