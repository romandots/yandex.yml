package render

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"
	"yandex-export/config"
	"yandex-export/entity"
	"yandex-export/repository"
)

var currentVersion entity.Version
var mu = &sync.Mutex{}

// XmlHandler генерирует YML и отдаёт его в ответе
func XmlHandler(w http.ResponseWriter, sr *http.Request) {
	var passLink, classLink string
	params := sr.URL.Query()
	if len(params) > 0 {
		passLink = params.Get("passlink")
		classLink = params.Get("classlink")
	}

	if passLink != "" {
		config.PassDefaultLink = passLink
	}

	if classLink != "" {
		config.ClassDefaultLink = classLink
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

	hash := HashOffers(offers)
	mu.Lock()
	if currentVersion.Hash != hash {
		currentVersion.PubDate = time.Now().Format("2006-01-02T15:04-07:00")
		currentVersion.Hash = hash
		log.Println("Updating version: " + currentVersion.PubDate)
	}
	mu.Unlock()

	catalogWithDate := entity.YmlCatalog{
		Name:    config.CompanyName,
		Company: config.CompanyName,
		Date:    currentVersion.PubDate,
		Shop: entity.Shop{
			Categories: config.Categories,
			Offers:     entity.Offers{Offer: offers},
		},
	}

	outputWithDate, err := xml.MarshalIndent(catalogWithDate, "", "  ")
	if err != nil {
		http.Error(w, fmt.Sprintf("XML marshal error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Write([]byte(xml.Header))
	w.Write(outputWithDate)
}

func HashBytes(b []byte) string {
	hash := sha256.Sum256(b)
	return hex.EncodeToString(hash[:])
}

// HashOffers creates a hash based on the offers data from database
func HashOffers(offers []entity.Offer) string {
	// Sort offers by ID to ensure consistent hashing
	sort.Slice(offers, func(i, j int) bool {
		return offers[i].ID < offers[j].ID
	})

	hasher := sha256.New()

	for _, offer := range offers {
		// Include key fields that represent the data state
		fmt.Fprintf(hasher, "%d|%s|%s|%s|%d|%d|%s|%s|",
			offer.ID,
			offer.Name,
			offer.Description,
			offer.Vendor,
			offer.Price,
			offer.CategoryID,
			offer.CurrencyID,
			offer.URL,
		)

		// Include ShortDescription if it exists
		if offer.ShortDescription != "" {
			fmt.Fprintf(hasher, "%s|", offer.ShortDescription)
		}
	}

	return hex.EncodeToString(hasher.Sum(nil))
}
