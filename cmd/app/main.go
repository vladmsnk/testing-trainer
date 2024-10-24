package main

import (
	"database/sql"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"log"
	"strconv"
	"testing_trainer/internal/usecase/progress"
	"testing_trainer/scripts/migrations"

	"testing_trainer/config"
	"testing_trainer/internal/storage"
	"testing_trainer/internal/usecase/habit"
	"testing_trainer/internal/usecase/user"
)

func main() {
	defer func() {
		log.Print(recover())
	}()

	err := config.Init()
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

	// usecases
	var (
		authUc    = user.New(store)
		habitUc   = habit.New(store, authUc)
		processUc = progress.New(authUc, store)
	)

	router := setupRouter(authUc, habitUc, processUc)
	log.Println("Swagger is available on http://" + config.ConfigStruct.HTTP.Host + ":" + strconv.Itoa(config.ConfigStruct.HTTP.Port) + "/swagger/index.html")
	err = router.Run(config.ConfigStruct.HTTP.Host + ":" + strconv.Itoa(config.ConfigStruct.HTTP.Port))
	if err != nil {
		log.Fatal(err)
	}
}

func GetSqlDBFromPgxPool(pool *pgxpool.Pool) *sql.DB {
	return stdlib.OpenDB(*pool.Config().ConnConfig)
}
