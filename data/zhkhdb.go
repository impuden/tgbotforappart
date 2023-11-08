package data

import (
	"database/sql"
)

func GetZhkhAmount(db *sql.DB) (int, error) {
	var zhkhAmount int
	err := db.QueryRow("SELECT value FROM expenses WHERE name = 'ЖКХ'").Scan(&zhkhAmount)
	if err != nil {
		return 0, err
	}

	return zhkhAmount, nil
}

func Writezhkh(db *sql.DB, chatID int64, value float64) error {
	_, err := db.Exec(`
        INSERT INTO expenses (name, value)
        VALUES ($1, $2)
        ON CONFLICT (name) DO UPDATE
        SET value = EXCLUDED.value
    `, "ЖКХ", value)
	return err
}

func Delzhkh(db *sql.DB, chatID int64) error {
	_, err := db.Exec(`
		INSERT INTO expenses (name, value)
		VALUES ($1, $2)
		ON CONFLICT (name) DO UPDATE
		SET value = 0.0
	`, "ЖКХ", 0.0)
	return err
}
