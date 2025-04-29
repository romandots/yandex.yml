package main

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"os"
)

// Обёртка для CDATA
type CData struct {
	Text string `xml:",cdata"`
}

type YmlCatalog struct {
	XMLName xml.Name `xml:"yml_catalog"`
	Shop    Shop     `xml:"shop"`
}

type Shop struct {
	Categories Categories `xml:"categories"`
	Offers     Offers     `xml:"offers"`
}

type Categories struct {
	Category []Category `xml:"category"`
}

type Category struct {
	ID   int    `xml:"id,attr"`
	Name string `xml:",chardata"`
}

type Offers struct {
	Offer []Offer `xml:"offer"`
}

type Offer struct {
	XMLName     xml.Name `xml:"offer"`
	ID          int      `xml:"id,attr"`
	Vendor      string   `xml:"vendor"`
	Price       int      `xml:"price"`
	CurrencyID  string   `xml:"currencyId"`
	CategoryID  int      `xml:"categoryId"`
	Picture     string   `xml:"picture"`
	URL         string   `xml:"url"`
	Name        string   `xml:"name"`
	Description *CData   `xml:"description"`
}

var db *sql.DB

func main() {
	fmt.Println("→ [1] старт программы")

	// === 1. Настройка подключения ===
	dsn := "admin_bezpravil:90ung11gTv@tcp(localhost:3306)/admin_krd_bezpravil?charset=utf8mb4&parseTime=true"
	var err error
	db, err = sql.Open("mysql", dsn)
	fmt.Println("→ [2] sql.Open завершился")

	if err != nil {
		log.Fatalf("DB connect error: %v", err)
	}
	defer db.Close()
	fmt.Println("→ [2] sql.Open завершился")

	// Проверим соединение
	if err := db.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "db.Ping error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("→ [3] ping к БД успешен")

	http.HandleFunc("/yandex_services.xml", xmlHandler)
	port := getPort()
	fmt.Printf("→ [4] слушаем на %s\n", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "ListenAndServe error: %v\n", err)
		os.Exit(1)
	}
}

// getPort берёт порт из env, или ":8080" по умолчанию
func getPort() string {
	if p := os.Getenv("PORT"); p != "" {
		return ":" + p
	}
	return ":9999"
}

// xmlHandler генерирует YML и отдаёт его в ответе
func xmlHandler(w http.ResponseWriter, r *http.Request) {
	offers, err := fetchOffers()
	if err != nil {
		http.Error(w, fmt.Sprintf("fetchOffers error: %v", err), http.StatusInternalServerError)
		return
	}

	// Категория — статично
	cats := []Category{
		{ID: 101, Name: "Танцевальные классы"},
	}

	catalog := YmlCatalog{
		Shop: Shop{
			Categories: Categories{Category: cats},
			Offers:     Offers{Offer: offers},
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

// fetchOffers тянет из БД текущие записи из classes
func fetchOffers() ([]Offer, error) {
	query := `
        SELECT id, ` + "`string`" + ` AS name, description
        FROM classes
        WHERE (end_date IS NULL OR end_date < NOW())
          AND (start_date IS NULL OR start_date > NOW())
          AND hidden IS NULL
          AND deleted IS NULL
    `
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Offer
	for rows.Next() {
		var (
			o    Offer
			desc sql.NullString
		)
		// Scan: id, name, desc
		if err := rows.Scan(&o.ID, &o.Name, &desc); err != nil {
			return nil, err
		}
		o.Vendor = "Школа танцев «Без правил»"
		o.Price = 4300
		o.CurrencyID = "RUR"
		o.CategoryID = 101
		o.Picture = ""
		o.URL = ""
		if desc.Valid {
			o.Description = &CData{Text: desc.String}
		}
		list = append(list, o)
	}
	return list, rows.Err()
}
