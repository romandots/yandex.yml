package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"yandex-export/config"
	"yandex-export/render"
)

func InitAndRun() {
	http.HandleFunc(config.YandexPath, render.XmlHandler)
	port := ":" + config.Port
	log.Printf("Слушаем порт %s\n", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "ListenAndServe error: %v\n", err)
		os.Exit(1)
	}
}
