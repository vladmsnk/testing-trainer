package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"log"
	"testing_trainer/cmd/docs"
	"testing_trainer/config"
	"testing_trainer/internal/app/auth"
	"testing_trainer/internal/app/habit"
	"testing_trainer/internal/app/progress"
	"testing_trainer/internal/app/time_swticher"
	"testing_trainer/internal/usecase/goals_checker"
	"testing_trainer/internal/usecase/progress_adder"
	"testing_trainer/internal/usecase/progress_getter"
	"testing_trainer/internal/usecase/progress_recalculator"
	"testing_trainer/internal/usecase/time_switcher"
	"testing_trainer/internal/usecase/user"
	"testing_trainer/middlewares"
	"time"
)

func setupRouter(userUc user.UseCase, habitUc habit.UseCase, progressUc progress_adder.UseCase, progressGetter progress_getter.ProgressGetter, progressRecalculator progress_recalculator.UseCase, timeSwitcher time_switcher.UseCase) *gin.Engine {
	r := gin.Default()

	docs.SwaggerInfo.BasePath = "/api/v1"

	r.GET("/api/v1/version", getVersion)

	authHandlers := r.Group("/api/v1/auth")
	auth.NewAuthHandler(authHandlers, userUc)

	protectedHabitHandlers := r.Group("/api/v1/tracker")
	protectedHabitHandlers.Use(middlewares.AuthMiddleware(userUc))

	protectedTimeSwitchHandlers := r.Group("/api/v1/time")
	protectedTimeSwitchHandlers.Use(middlewares.AuthMiddleware(userUc))

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	habit.NewHabitHandler(protectedHabitHandlers, habitUc)
	progress.NewProgressHandler(protectedHabitHandlers, progressUc, progressGetter)
	time_swticher.NewHandler(protectedTimeSwitchHandlers, timeSwitcher)
	return r
}

func initPostgreSQLConnection(config config.Postgres) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), config.GetConnectionString())
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New: %w", err)
	}
	return pool, nil
}

func runCheckGoalsScheduler(checker goals_checker.Checker) (gocron.Scheduler, error) {
	scheduler, err := gocron.NewScheduler(gocron.WithLocation(time.UTC))
	if err != nil {
		return nil, fmt.Errorf("gocron.NewScheduler: %w", err)
	}

	cron := gocron.CronJob("1 * * * *", false)
	task := gocron.NewTask(func() {
		err := checker.CheckGoals(context.Background())
		if err != nil {
			log.Println("Error while checking goals: ", err)
		}
		log.Println("Goals checked")
		return
	})

	_, err = scheduler.NewJob(cron, task)
	if err != nil {
		return nil, fmt.Errorf("scheduler.NewJob: %w", err)
	}

	return scheduler, nil
}

// getVersion godoc
// @Summary Get API version
// @Description Get the current version of the API
// @Tags version
// @Accept json
// @Produce json
// @Success 200 {object} VersionResponse "Successful response"
// @Router /version [get]
func getVersion(c *gin.Context) {
	c.JSON(200, VersionResponse{Version: "1.0.0"})
}

type VersionResponse struct {
	Version string `json:"version" example:"1.0.0"`
}
