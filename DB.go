package main

import (
	"database/sql"
	"errors"
	"math"
	"time"
)

func initializeDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "requests.db")
	if err != nil {
		return nil, err
	}

	createPagesTable := `
        CREATE TABLE IF NOT EXISTS pages (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            url TEXT,
            created_at TIMESTAMP
        );
    `
	createRequestsTable := `
        CREATE TABLE IF NOT EXISTS requests (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            page_id INT,
            status INT,
            elapsed_time REAL,
            created_at TIMESTAMP
        );
    `
	_, err = db.Exec(createPagesTable)
	if err != nil {
		log.Error("Ошибка при создании таблицы pages:", err)
		return nil, err
	}
	_, err = db.Exec(createRequestsTable)
	if err != nil {
		log.Error("Ошибка при создании таблицы requests:", err)
		return nil, err
	}

	return db, nil
}

func saveRequestDB(db *sql.DB, pageID int64, status int, elapsedTime time.Duration) {
	roundedElapsedTime := math.Round(elapsedTime.Seconds()*1000) / 1000
	insertSQL := "INSERT INTO requests (page_id, status, elapsed_time, created_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)"
	_, err := db.Exec(insertSQL, pageID, status, roundedElapsedTime)
	if err != nil {
		log.Error("Ошибка при сохранении данных:", err)
	}
}

func savePageDB(db *sql.DB, url string) (int64, error) {
	var pageID int64
	err := db.QueryRow("SELECT id FROM pages WHERE url = ?", url).Scan(&pageID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	if pageID != 0 {
		return pageID, nil
	}

	result, err := db.Exec("INSERT INTO pages (url, created_at) VALUES (?, CURRENT_TIMESTAMP)", url)
	if err != nil {
		return 0, err
	}
	pageID, err = result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return pageID, nil
}
