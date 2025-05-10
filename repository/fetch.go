package repository

import (
	"database/sql"
	"fmt"
	"log"
	"os"
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

	if err := db.Ping(); err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
		os.Exit(1)
	}
	log.Println("Подключились к БД")

	return db, err
}

// FetchClasses тянет из БД текущие записи из classes
func FetchClasses() ([]entity.Offer, error) {
	query := `
		SELECT
  c.id,
  c.string       AS name,
  c.description  AS class_description,
  (
    SELECT st.description
    FROM styles_classes AS sc
    JOIN styles          AS st ON st.id = sc.style_id
    WHERE sc.class_id = c.id
    ORDER BY sc.id DESC, st.id DESC
    LIMIT 1
  ) AS style_description,
  c.mon, c.tue, c.wed, c.thu, c.fri, c.sat, c.sun,
  s.studio_title,
		c.price_rate
FROM classes AS c
JOIN studios AS s
  ON c.studio_id = s.id
WHERE c.hidden   IS NULL
  AND c.deleted  IS NULL
  AND c.string   IS NOT NULL
  AND (c.start_date IS NULL OR c.start_date <= NOW())
  AND (c.end_date   IS NULL OR c.end_date   >= NOW());
    `

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []entity.Offer
	for rows.Next() {
		o, err := scanClass(rows)
		if err != nil {
			return list, err
		}
		list = append(list, o)
	}
	return list, rows.Err()
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
		{
			ID:          1,
			Name:        "Первое пробное занятие",
			Description: "Первый урок в любом классе",
			Vendor:      config.CompanyName,
			Price:       config.FirstVisitPrice,
			CurrencyID:  "RUR",
			CategoryID:  2,
			Picture:     config.PassDefaultPicture,
			URL:         config.PassDefaultLink,
		},
		{
			ID:          2,
			Name:        "Разовое занятие",
			Description: "Одно часовое посещение в любом классе",
			Vendor:      config.CompanyName,
			Price:       config.VisitPrice,
			CurrencyID:  "RUR",
			CategoryID:  2,
			Picture:     config.PassDefaultPicture,
			URL:         config.PassDefaultLink,
		},
	}
	var id int = 3
	for rows.Next() {
		id++
		o, empty, err := scanPass(rows, id)
		if err != nil {
			return list, err
		}
		if empty {
			continue
		}
		list = append(list, o)
	}
	return list, rows.Err()
}

func scanClass(rows *sql.Rows) (entity.Offer, error) {
	var (
		o         entity.Offer
		name      string
		classDesc sql.NullString
		styleDesc sql.NullString
		mon       sql.NullString
		tue       sql.NullString
		wed       sql.NullString
		thu       sql.NullString
		fri       sql.NullString
		sat       sql.NullString
		sun       sql.NullString
		studio    sql.NullString
		price     sql.NullInt64
	)
	if err := rows.Scan(
		&o.ID, &name, &classDesc, &styleDesc, &mon, &tue, &wed, &thu, &fri, &sat, &sun, &studio, &price,
	); err != nil {
		return entity.Offer{}, err
	}

	if studio.Valid {
		name += " в студии " + studio.String
	}

	var description string
	schedule := getSchedule(mon, tue, wed, thu, fri, sat, sun)
	if classDesc.Valid && classDesc.String != "" {
		description = classDesc.String + "\n"
	} else if styleDesc.Valid && styleDesc.String != "" {
		description = styleDesc.String + "\n"
	}

	fullDescription := description + schedule
	shortDescription := common.SafelyTruncate(schedule, 250)

	o.Name = common.SafelyTruncate(name, 250)
	o.Vendor = config.CompanyName
	o.Description = fullDescription
	if len(fullDescription) > 250 {
		o.ShortDescription = shortDescription
	} else {
		o.ShortDescription = fullDescription
	}
	if price.Valid {
		o.Price = int(price.Int64)
	} else {
		o.Price = config.VisitPrice
	}
	o.Picture = config.ClassDefaultPicture
	o.URL = config.ClassDefaultLink
	o.CurrencyID = "RUR"
	o.CategoryID = 1
	return o, nil
}

func scanPass(rows *sql.Rows, id int) (entity.Offer, bool, error) {
	var (
		o              entity.Offer
		name           string
		desc           sql.NullString
		price          sql.NullInt64
		lifetime       sql.NullInt64
		hours          sql.NullInt64
		freeze_allowed sql.NullInt64
		guest_visits   sql.NullInt64
	)
	if err := rows.Scan(
		&name, &desc, &price,
		&lifetime, &hours, &freeze_allowed, &guest_visits,
	); err != nil {
		return entity.Offer{}, false, err
	}

	if !price.Valid || !desc.Valid || !lifetime.Valid || !hours.Valid {
		return o, true, nil
	}

	var freezeAllowed string
	if freeze_allowed.Valid && freeze_allowed.Int64 > 0 {
		freezeAllowed = "C возможностью заморозки на месяц."
	}

	var guestVisits string
	if guest_visits.Valid && guest_visits.Int64 > 0 {
		guestVisits = fmt.Sprintf(" + %s для друзей",
			common.Inflect(int(guest_visits.Int64), []string{"гостевое", "гостевых", "гостевых"}),
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
		lifetimeString = fmt.Sprintf(" на %s. ",
			common.Inflect(int(lifetime.Int64), []string{"день", "дня", "дней"}),
		)
	}

	fullDescription := desc.String + lifetimeString + lessonsIncluded + freezeAllowed
	shortDescription := common.SafelyTruncate(desc.String, 250)

	o.ID = id
	o.Name = common.SafelyTruncate(name, 250)
	o.Description = fullDescription
	if len(fullDescription) > 250 {
		o.ShortDescription = shortDescription
	} else {
		o.ShortDescription = fullDescription
	}
	o.Vendor = config.CompanyName
	o.Price = int(price.Int64)
	o.Picture = config.PassDefaultPicture
	o.URL = config.PassDefaultLink
	o.CurrencyID = "RUR"
	o.CategoryID = 2
	return o, false, nil
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
		str += " в " + time

		scheduleStrings = append(scheduleStrings, str)
	}

	return strings.Join(scheduleStrings, "; ")
}
