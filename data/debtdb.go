package data

import (
	"database/sql"
	"log"
)

func Writedebt(db *sql.DB) {
	people := []string{"Арсений", "Максим", "Егор", "Владимир"}

	for _, person := range people {
		var value float64
		err := db.QueryRow("SELECT value FROM expenses WHERE name = $1", person).Scan(&value)
		if err != nil {
			log.Printf("Failed to get value for %s: %v\n", person, err)
			continue
		}

		value /= 4 // Делим на 4

		switch person {
		case "Арсений":
			_, err = db.Exec("UPDATE debt SET ars = $1", value)
		case "Максим":
			_, err = db.Exec("UPDATE debt SET max = $1", value)
		case "Егор":
			_, err = db.Exec("UPDATE debt SET egor = $1", value)
		case "Владимир":
			_, err = db.Exec("UPDATE debt SET vova = $1", value)
		default:
			log.Println("Unknown person")
		}

		if err != nil {
			log.Printf("Failed to update debt for %s: %v\n", person, err)
		}
	}
}

func Readdebt(db *sql.DB, name string) (float64, float64, float64, float64, error) {
	var (
		ars  float64
		max  float64
		egor float64
		vova float64
	)
	err := db.QueryRow("SELECT ars, max, egor, vova FROM debt WHERE name = $1", name).Scan(&ars, &max, &egor, &vova)
	if err != nil {
		log.Println("fail by exec debt", err)
		return 0, 0, 0, 0, err
	}

	return ars, max, egor, vova, nil
}

func Deletedebt(db *sql.DB, name string, sname string) error {
	// Маппинг имен на их сокращения
	nameMapping := map[string]string{
		"Арсений":  "ars",
		"Владимир": "vova",
		"Егор":     "egor",
		"Максим":   "max",
	}

	// Проверка наличия соответствующего сокращения имени в мапе
	shortenedName, ok := nameMapping[sname]
	if !ok {
		log.Printf("Неправильное сокращение имени: %s", sname)
		return nil
	}

	query := "UPDATE debt SET " + shortenedName + " = 0 WHERE name = $1"

	_, err := db.Exec(query, name)
	if err != nil {
		log.Printf("Ошибка при выполнении запроса к БД: %v", err)
		return err
	}

	log.Printf("Данные обновлены успешно для: %s", name)
	return nil
}
