package data

import (
	"database/sql"
	"strings"
)

func Writeexp(db *sql.DB, name string, value float64, comment string) error {
	// Получаем текущие значения для указанного имени
	row := db.QueryRow("SELECT value, comment FROM expenses WHERE name = $1", name)

	var (
		currentValue   *float64
		currentComment *string
	)

	// Сканируем текущие значения из базы данных
	err := row.Scan(&currentValue, &currentComment)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// Проверяем текущие значения на NULL и заменяем их, если они NULL или пустые строки
	if currentValue == nil {
		currentValue = new(float64)
	}
	if currentComment == nil || *currentComment == "" {
		currentComment = new(string)
	}

	// Проверяем текущий комментарий на пустую строку или наличие в нем только пробелов
	if strings.TrimSpace(*currentComment) != "" {
		// Добавляем запятую, если комментарий уже содержит данные
		comment = ", " + comment
	}

	// Добавляем новые значения к существующим
	*currentValue += value
	*currentComment += comment

	// Обновляем запись в базе данных
	_, err = db.Exec(`
        INSERT INTO expenses (name, value, comment)
        VALUES ($1, $2, $3)
        ON CONFLICT (name) DO UPDATE
        SET value = $4,
            comment = $5;
    `, name, *currentValue, *currentComment, *currentValue, *currentComment)

	return err
}

func Getexp(db *sql.DB, name string) (float64, string, error) {
	var value float64
	var comment string
	err := db.QueryRow("SELECT value, comment FROM expenses WHERE name = $1", name).Scan(&value, &comment)
	if err != nil {
		return 0, "", err
	}

	return value, comment, nil
}

func Delexp(db *sql.DB, name string) error {
	_, err := db.Exec("UPDATE expenses SET value = 0, comment = '' WHERE name = $1", name)
	return err
}
