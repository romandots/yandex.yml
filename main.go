package main

import (
	_ "github.com/go-sql-driver/mysql"
	"yandex-export/repository"
	"yandex-export/server"
)

func main() {
	db, err := repository.InitDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	server.InitAndRun()
}
