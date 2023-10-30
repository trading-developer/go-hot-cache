package main

import (
	"database/sql"
	"fmt"
	"time"
)

func saveDB(db *sql.DB, apiURL string, status int, elapsedTime time.Duration) {
	insertSQL := "INSERT INTO requests (url, status, elapsed_time, created_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)"
	_, err := db.Exec(insertSQL, apiURL, status, elapsedTime.Seconds())
	if err != nil {
		fmt.Println("Ошибка при сохранении данных в SQLite:", err)
	}
}

func initializeDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "requests.db")
	if err != nil {
		return nil, err
	}

	createTableSQL := `
        CREATE TABLE IF NOT EXISTS requests (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            url TEXT,
            status INT,
            elapsed_time REAL,
            created_at TIMESTAMP
        );
    `

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, err
	}

	return db, nil
}
