package render

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"time"
	"yandex-export/config"
	"yandex-export/entity"
	"yandex-export/repository"
)

// XmlHandler генерирует YML и отдаёт его в ответе
func XmlHandler(w http.ResponseWriter, sr *http.Request) {
	var passLink, classLink string
	params := sr.URL.Query()
	if len(params) > 0 {
		passLink = params.Get("passlink")
		classLink = params.Get("classlink")
	}

	if passLink == "" {
		passLink = config.PassDefaultLink
	}

	if classLink == "" {
		classLink = config.ClassDefaultLink
	}

	log.Println("passLink:", passLink)
	log.Println("classLink:", classLink)

	classes, err := repository.FetchClasses()
	if err != nil {
		http.Error(w, fmt.Sprintf("fetchClasses error: %v", err), http.StatusInternalServerError)
		return
	}

	passes, err := repository.FetchPasses()
	if err != nil {
		http.Error(w, fmt.Sprintf("fetchPasses error: %v", err), http.StatusInternalServerError)
		return
	}

	offers := make([]entity.Offer, 0, len(classes)+len(passes))
	offers = append(offers, classes...)
	offers = append(offers, passes...)

	catalog := entity.YmlCatalog{
		Date:    time.Now().Format("2006-01-02T15:04-07:00"),
		Name:    config.CompanyName,
		Company: config.CompanyName,
		Shop: entity.Shop{
			Categories: config.Categories,
			Offers:     entity.Offers{Offer: offers},
		},
	}

	output, err := xml.MarshalIndent(catalog, "", "  ")
	if err != nil {
		http.Error(w, fmt.Sprintf("XML marshal error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Write([]byte(xml.Header))
	w.Write(output)
}
