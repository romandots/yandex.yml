package repository

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
	"yandex-export/common"
	"yandex-export/config"
	"yandex-export/entity"
)

var db *sql.DB

func InitDB() (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4",
		config.DatabaseConfig.User,
		config.DatabaseConfig.Password,
		config.DatabaseConfig.Host,
		config.DatabaseConfig.Port,
		config.DatabaseConfig.DBName,
	)
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	log.Println("Подключились к БД")

	return db, err
}

// FetchPasses тянет из БД текущие записи из passes
func FetchPasses() ([]entity.Offer, error) {
	query := `
		SELECT t.ticket_type_name                      AS name,
			   t.description,
			   t.default_price                         AS price,
			   t.default_period                        AS lifetime,
			   CAST(t.default_periods / 2 AS UNSIGNED) AS hours,
			   t.default_frosts AS freeze_allowed,
			   t.default_guests AS guest_visits
		FROM ticket_types AS t
		WHERE t.ticket_type_active = 1 AND t.description IS NOT NULL
		ORDER BY t.default_price ASC
`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []entity.Offer = []entity.Offer{
		{ID: 1, Name: "Первое пробное занятие", Vendor: config.CompanyName, Price: config.FirstVisitPrice, CurrencyID: "RUR", CategoryID: 2},
		{ID: 2, Name: "Разовое занятие", Vendor: config.CompanyName, Price: config.VisitPrice, CurrencyID: "RUR", CategoryID: 2},
	}
	var id int = 3
	for rows.Next() {
		id++
		var (
			o              entity.Offer
			desc           sql.NullString
			price          sql.NullInt64
			lifetime       sql.NullInt64
			hours          sql.NullInt64
			freeze_allowed sql.NullInt64
			guest_visits   sql.NullInt64
		)
		if err := rows.Scan(
			&o.Name, &desc, &price,
			&lifetime, &hours, &freeze_allowed, &guest_visits,
		); err != nil {
			return nil, err
		}

		if !price.Valid || !desc.Valid || !lifetime.Valid || !hours.Valid {
			continue
		}

		var freezeAllowed string
		if freeze_allowed.Valid && freeze_allowed.Int64 > 0 {
			freezeAllowed = "Доступна «заморозка» абонемента. "
		}

		var guestVisits string
		if guest_visits.Valid && guest_visits.Int64 > 0 {
			guestVisits = fmt.Sprintf(" + %s для друзей",
				common.Inflect(int(guest_visits.Int64), []string{"гостевое посещение", "гостевых посещения", "гостевых посещений"}),
			)
		}

		var lessonsIncluded string
		if hours.Int64 > 0 {
			lessonsIncluded = fmt.Sprintf("Включено %s%s. ",
				common.Inflect(int(hours.Int64), []string{"урок", "урока", "уроков"}),
				guestVisits,
			)
		}

		var lifetimeString string
		if lifetime.Int64 > 0 {
			lifetimeString = fmt.Sprintf("Срок действия: %s. ",
				common.Inflect(int(lifetime.Int64), []string{"день", "дня", "дней"}),
			)
		}

		o.Description = &entity.CData{Text: desc.String + ".\n" + lifetimeString + lessonsIncluded + freezeAllowed}

		o.Vendor = config.CompanyName
		o.CurrencyID = "RUR"
		o.CategoryID = 2
		o.Price = int(price.Int64)
		o.ID = id
		o.Picture = ""
		o.URL = ""
		list = append(list, o)
	}
	return list, rows.Err()
}

// FetchClasses тянет из БД текущие записи из classes
func FetchClasses() ([]entity.Offer, error) {
	query := `
		SELECT c.id, c.string AS name, c.description, c.mon, c.tue, c.wed, c.thu, c.fri, c.sat, c.sun
		FROM classes AS c
				 JOIN (
			SELECT string, MIN(id) AS min_id
			FROM classes
			WHERE (start_date IS NULL OR start_date <= NOW())
			  AND (end_date IS NULL OR end_date >= NOW())
			  AND hidden IS NULL
			  AND deleted IS NULL
			GROUP BY string
		) AS sub ON sub.min_id = c.id AND c.string = sub.string
    `

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []entity.Offer
	for rows.Next() {
		var (
			o    entity.Offer
			desc sql.NullString
			mon  sql.NullString
			tue  sql.NullString
			wed  sql.NullString
			thu  sql.NullString
			fri  sql.NullString
			sat  sql.NullString
			sun  sql.NullString
		)
		if err := rows.Scan(
			&o.ID, &o.Name, &desc, &mon, &tue, &wed, &thu, &fri, &sat, &sun,
		); err != nil {
			return nil, err
		}

		var description string
		if desc.Valid {
			description = desc.String
		}

		schedule := getSchedule(mon, tue, wed, thu, fri, sat, sun)
		if schedule != "" {
			description += "\n" + schedule
		}

		o.Description = &entity.CData{Text: description}
		o.Vendor = config.CompanyName
		o.Price = config.BasicPassPrice
		o.CurrencyID = "RUR"
		o.CategoryID = 1
		o.Picture = ""
		o.URL = ""
		list = append(list, o)
	}
	return list, rows.Err()
}

func getSchedule(mon sql.NullString, tue sql.NullString, wed sql.NullString, thu sql.NullString, fri sql.NullString, sat sql.NullString, sun sql.NullString) string {
	days := map[int]string{
		0: "понедельникам",
		1: "вторникам",
		2: "средам",
		3: "четвергам",
		4: "пятницам",
		5: "субботам",
		6: "воскресеньям",
	}

	schedules := make(map[string][]string)
	for dayIndex, timeValue := range []sql.NullString{mon, tue, wed, thu, fri, sat, sun} {
		if timeValue.Valid {
			day := days[dayIndex]
			timeString := timeValue.String
			if time, err := time.Parse("15:04:04", timeString); err == nil {
				timeString := time.Format("15:04")
				if schedules[timeString] == nil {
					schedules[timeString] = make([]string, 0)
				}
				schedules[timeString] = append(schedules[timeString], day)
			}
		}
	}

	scheduleStrings := make([]string, 0, len(schedules))
	for time, days := range schedules {
		str := "По "
		var lastDay string
		if len(days) > 1 {
			lastDay = days[len(days)-1]
			days = days[:len(days)-1]
			str += strings.Join(days, ", ") + " и " + lastDay
		} else {
			str += days[0]
		}
		str += " в " + time + "."

		scheduleStrings = append(scheduleStrings, str)
	}

	return strings.Join(scheduleStrings, "\n")
}
