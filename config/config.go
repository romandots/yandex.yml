package config

import (
	"github.com/joho/godotenv"
	"log"
	"yandex-export/common"
	"yandex-export/entity"
)

var CompanyName = "Школа танцев «Без правил»"
var FirstVisitPrice int
var VisitPrice int

var Categories entity.Categories = entity.Categories{[]entity.Category{
	{ID: 1, Name: "Танцевальные классы (разовое посещение)"},
	{ID: 2, Name: "Абонементы"},
}}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

var DatabaseConfig *DBConfig
var Port string
var YandexPath string
var LogoUrl string
var CompanyUrl string

func init() {
	// Попробуем загрузить .env из текущей папки.
	// Если файла нет — продолжаем без фатальной ошибки.
	if err := godotenv.Load(); err != nil {
		log.Println(".env файл отсутствует:", err)
	} else {
		log.Println("Загрузили конфиг из .env файла")
	}

	FirstVisitPrice = common.GetEnvInt("FIRST_VISIT_PRICE", 300)
	VisitPrice = common.GetEnvInt("VISIT_PRICE", 700)
	DatabaseConfig = &DBConfig{
		Host:     common.GetEnvString("DB_HOST", "localhost"),
		Port:     common.GetEnvString("DB_PORT", "3306"),
		User:     common.GetEnvString("DB_USER", "root"),
		Password: common.GetEnvString("DB_PASSWORD", ""),
		DBName:   common.GetEnvString("DB_NAME", "root"),
	}
	Port = common.GetEnvString("PORT", "9999")
	YandexPath = common.GetEnvString("YANDEX_PATH", "/yandex.yml")
	LogoUrl = common.GetEnvString("LOGO_URL", "https://bezpravil.net/img/logo.png")
	CompanyUrl = common.GetEnvString("COMPANY_URL", "https://bezpravil.net")
}
