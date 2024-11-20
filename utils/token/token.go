package token

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	keyEnvTokenLifeSpan   = "TOKEN_HOUR_LIFESPAN"
	keyEnvRefreshLifeSpan = "REFRESH_HOUR_LIFESPAN"
	KeyEnvApiSecret       = "API_SECRET"
	KeyEnvRefreshSecret   = "REFRESH_SECRET"
)

var (
	ErrTokenExpired = fmt.Errorf("token has expired")
	ErrInvalidToken = fmt.Errorf("invalid token")
)

func ExtractToken(c *gin.Context) string {
	bearerToken := c.Request.Header.Get("Authorization")
	if len(strings.Split(bearerToken, " ")) == 2 {
		return strings.Split(bearerToken, " ")[1]
	}
	return ""
}

func ExtractUsernameFromToken(c *gin.Context) (string, error) {
	tokenString := ExtractToken(c)

	username, err := ExtractUserNameFromToken(tokenString, KeyEnvApiSecret)
	if err != nil {
		return "", fmt.Errorf("failed to extract username from token: %w", err)
	}

	return username, nil
}

func GenerateTokens(username string) (string, string, error) {
	accessTokenLifespanStr := os.Getenv(keyEnvTokenLifeSpan)
	if accessTokenLifespanStr == "" {
		return "", "", fmt.Errorf("TOKEN_HOUR_LIFESPAN environment variable not set")
	}

	accessTokenLifespan, err := strconv.Atoi(accessTokenLifespanStr)
	if err != nil {
		return "", "", fmt.Errorf("invalid TOKEN_HOUR_LIFESPAN: %w", err)
	}

	accessClaims := jwt.MapClaims{
		"authorized": true,
		"username":   username,
		"exp":        time.Now().Add(time.Hour * time.Duration(accessTokenLifespan)).Unix(), // Access token expiry
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	apiSecret := os.Getenv(KeyEnvApiSecret)
	if apiSecret == "" {
		return "", "", fmt.Errorf("API_SECRET environment variable not set")
	}

	accessTokenString, err := accessToken.SignedString([]byte(apiSecret))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign access token: %w", err)
	}

	refreshTokenLifespanStr := os.Getenv(keyEnvRefreshLifeSpan)
	if refreshTokenLifespanStr == "" {
		return "", "", fmt.Errorf("REFRESH_HOUR_LIFESPAN environment variable not set")
	}

	refreshTokenLifespan, err := strconv.Atoi(refreshTokenLifespanStr)
	if err != nil {
		return "", "", fmt.Errorf("invalid REFRESH_HOUR_LIFESPAN: %w", err)
	}

	refreshClaims := jwt.MapClaims{
		"authorized": true,
		"username":   username,
		"exp":        time.Now().Add(time.Hour * time.Duration(refreshTokenLifespan)).Unix(), // Refresh token expiry
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshSecret := os.Getenv(KeyEnvRefreshSecret)
	if refreshSecret == "" {
		return "", "", fmt.Errorf("REFRESH_SECRET environment variable not set")
	}

	refreshTokenString, err := refreshToken.SignedString([]byte(refreshSecret))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return accessTokenString, refreshTokenString, nil
}

func TokenValidInContext(c *gin.Context) error {
	tokenString := ExtractToken(c)

	err := ValidateToken(tokenString, KeyEnvApiSecret)
	if err != nil {
		return err
	}
	return nil
}

func ExtractUserNameFromToken(tokenString string, keyEnvSecret string) (string, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv(keyEnvSecret)), nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}
	if !token.Valid {
		return "", fmt.Errorf("invalid token")
	}
	username, ok := claims["username"].(string)
	if !ok {
		return "", fmt.Errorf("username claim is missing or not a string")
	}
	return username, nil
}

func ValidateToken(tokenString string, keyEnvSecret string) error {
	envSecret := os.Getenv(keyEnvSecret)

	secret := []byte(envSecret)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return ErrInvalidToken
	}

	if !token.Valid {
		return ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return fmt.Errorf("unable to parse token claims")
	}

	if exp, ok := claims["exp"].(float64); ok {
		expirationTime := time.Unix(int64(exp), 0)
		if time.Now().After(expirationTime) {
			return ErrTokenExpired
		}
	} else {
		return fmt.Errorf("token does not have an expiration time")
	}

	return nil
}
