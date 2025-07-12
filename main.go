package main

import (
	"math/rand"
	"time"
	"yandex-export/repository"
	"yandex-export/server"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Initialize random seed for image selection
	rand.Seed(time.Now().UnixNano())

	db, err := repository.InitDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	server.InitAndRun()
}
