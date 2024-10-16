package main

import (
	"context"
	"fmt"
	"testing_trainer/internal/app/progress"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"testing_trainer/cmd/docs"
	"testing_trainer/config"
	"testing_trainer/internal/app/auth"
	"testing_trainer/internal/app/habit"
	"testing_trainer/internal/usecase/user"
	"testing_trainer/middlewares"
)

func setupRouter(userUc user.UseCase, habitUc habit.UseCase, progressUc progress.UseCase) *gin.Engine {
	r := gin.Default()

	docs.SwaggerInfo.BasePath = "/api/v1"
	authHandlers := r.Group("/api/v1/auth")
	auth.NewAuthHandler(authHandlers, userUc)

	protectedHabitHandlers := r.Group("/api/v1/tracker")
	protectedHabitHandlers.Use(middlewares.AuthMiddleware())

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	habit.NewHabitHandler(protectedHabitHandlers, habitUc)
	progress.NewProgressHandler(protectedHabitHandlers, progressUc)
	return r
}

func initPostgreSQLConnection(config config.Postgres) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), config.GetConnectionString())
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New: %w", err)
	}
	return pool, nil
}
