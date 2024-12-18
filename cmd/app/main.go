package main

import (
	"database/sql"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"log"
	"strconv"
	"testing_trainer/config"
	"testing_trainer/internal/storage"
	"testing_trainer/internal/storage/transactor"
	"testing_trainer/internal/usecase/goals_checker"
	"testing_trainer/internal/usecase/habit"
	"testing_trainer/internal/usecase/progress"
	"testing_trainer/internal/usecase/time_manager"
	"testing_trainer/internal/usecase/time_switcher"
	"testing_trainer/internal/usecase/user"
	"testing_trainer/scripts/migrations"
)

func main() {
	defer func() {
		log.Print(recover())
	}()

	err := config.InitConfigWithEnvs()
	if err != nil {
		log.Fatal(err)
	}

	pool, err := initPostgreSQLConnection(config.ConfigStruct.PG)
	if err != nil {
		log.Fatal(err)
	}

	db := GetSqlDBFromPgxPool(pool)
	// Apply migrations
	if err := migrations.ApplyMigrations(db); err != nil {
		log.Fatalf("failed to apply migrations: %v", err)
	}

	// storages
	var (
		store = storage.NewStorage(pool)
	)

	tx, err := transactor.New(pool)
	if err != nil {
		log.Fatal(err.Error())
	}

	var (
		timeManager = time_manager.New(store)

		authUc    = user.New(store)
		processUc = progress.New(authUc, store, tx, timeManager)

		habitUc        = habit.New(store, authUc, tx, timeManager, processUc)
		goalsCheckerUC = goals_checker.NewChecker(store, tx, timeManager, processUc)
		timeSwitcherUC = time_switcher.New(timeManager)
	)

	scheduler, err := runCheckGoalsScheduler(goalsCheckerUC)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		scheduler.Shutdown()
	}()
	scheduler.Start()

	router := setupRouter(authUc, habitUc, processUc, timeSwitcherUC)
	log.Println("Swagger is available on http://" + config.ConfigStruct.HTTP.Host + ":" + strconv.Itoa(config.ConfigStruct.HTTP.Port) + "/swagger/index.html")
	err = router.Run(config.ConfigStruct.HTTP.Host + ":" + strconv.Itoa(config.ConfigStruct.HTTP.Port))
	if err != nil {
		log.Fatal(err)
	}
}

func GetSqlDBFromPgxPool(pool *pgxpool.Pool) *sql.DB {
	return stdlib.OpenDB(*pool.Config().ConnConfig)
}
