package data

import (
	"database/sql"
	"fmt"
	"log"
	"tgappart/config"

	_ "github.com/lib/pq"
)

var db *sql.DB

func ConnectDB() (*sql.DB, error) {
	// Подключение к базе данных
	var err error
	db, err = sql.Open("postgres", config.ConnStr)
	if err != nil {
		log.Fatal(err)
		fmt.Println("no connection")
	}

	// Проверка подключения к базе данных
	err = db.Ping()
	if err != nil {
		fmt.Println("не ок")
	} else {
		fmt.Println("ок")
	}

	return db, nil
}
