package config

import (
	"log"
	"yandex-export/common"
	"yandex-export/entity"

	"github.com/joho/godotenv"
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
var ClassDefaultPicture string
var PassDefaultPicture string
var ClassDefaultLink string
var PassDefaultLink string
var ImageDir string
var ImagePath string

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
	ClassDefaultPicture = common.GetEnvString("CLASS_DEFAULT_PICTURE", "https://bezpravil.net/img/logo.png")
	ClassDefaultLink = common.GetEnvString("CLASS_DEFAULT_LINK", "https://bezpravil.net")
	PassDefaultPicture = common.GetEnvString("PASS_DEFAULT_PICTURE", "https://bezpravil.net/img/logo.png")
	PassDefaultLink = common.GetEnvString("PASS_DEFAULT_LINK", "https://bezpravil.net")
	ImageDir = common.GetEnvString("IMAGE_DIR", "images")
	ImagePath = common.GetEnvString("IMAGE_PATH", "https://bezpravil.net/img")
}
