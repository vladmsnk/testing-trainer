package main

import (
	"log"
	"strconv"

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

	strg := storage.NewStorage(pool)
	habitUc := habit.New(strg)
	authUc := user.New(strg)

	router := setupRouter(authUc, habitUc)
	err = router.Run(config.ConfigStruct.HTTP.Host + ":" + strconv.Itoa(config.ConfigStruct.HTTP.Port))
	if err != nil {
		log.Fatal(err)
	}
}
