package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"net/http"
	"sync"
	"time"
)

type Category struct {
	ID       int        `json:"id"`
	Slug     string     `json:"slug"`
	Children []Category `json:"children"`
}

var log = logrus.New()
var startTime = time.Now()

type PageInfo struct {
	URL         string
	ElapsedTime float64
	Status      int
}

func main() {
	setLogParam()
	readConfig()

	db, err := initializeDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	apiURL := "https://mageo.ru/api/v1/menu"

	response, err := http.Get(apiURL)
	if err != nil {
		fmt.Println("Ошибка при запросе API:", err)
		return
	}
	defer response.Body.Close()

	var categories []Category
	if err := json.NewDecoder(response.Body).Decode(&categories); err != nil {
		fmt.Println("Ошибка при декодировании JSON:", err)
		return
	}

	var wg sync.WaitGroup

	for _, category := range categories {
		wg.Add(1)
		go processCategory(db, category, &wg)
	}

	wg.Wait()

	sendTgStats(db)
}

func processCategory(db *sql.DB, category Category, wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Printf("ID: %d, Slug: %s\n", category.ID, category.Slug)

	apiURL := "https://mageo.ru/api/v1/category/" + category.Slug

	pageID, err := savePageDB(db, apiURL)
	if err != nil {
		fmt.Println("Ошибка при сохранении информации о странице:", err)
		log.Error("Ошибка при сохранении информации о странице:", err)
		return
	}

	start := time.Now()

	response, err := http.Get(apiURL)
	if err != nil {
		fmt.Println("Ошибка при запросе API:", err)
		log.Error("Ошибка при запросе API", err)
		return
	}
	defer response.Body.Close()

	elapsedTime := time.Since(start)

	saveRequestDB(db, pageID, response.StatusCode, elapsedTime)

	for _, childCategory := range category.Children {
		wg.Add(1)
		go processCategory(db, childCategory, wg)
	}
}

func sendTgStats(db *sql.DB) {

	topSlowPages := getTopSlowPages(db)
	errorsPages := getErrorsPages(db)

	if len(topSlowPages) == 0 && len(errorsPages) == 0 {
		fmt.Println("Нет медленных запросов и ошибок.")
		return
	}

	message := ""

	if len(topSlowPages) != 0 {
		message = "Топ-10 медленных запросов:\n"
		for i, page := range topSlowPages {
			message += fmt.Sprintf("%d. %s: %.2fсек.\n", i+1, page.URL, page.ElapsedTime)
		}
	}

	if len(errorsPages) != 0 {
		message += "Ошибки:\n"
		for i, page := range errorsPages {
			message += fmt.Sprintf("%d. %s: %d\n", i+1, page.URL, page.Status)
		}
	}

	bot, err := tgbotapi.NewBotAPI(appConfig.Token)
	if err != nil {
		log.Panic("Ошибка инициализации клиента tg:", err)
	}

	msg := tgbotapi.NewMessage(appConfig.ChatId, message)
	_, err = bot.Send(msg)
	if err != nil {
		log.Panic("Ошибка отправки сообщения в tg:", err)
	}
}

func getTopSlowPages(db *sql.DB) []PageInfo {
	query := `
SELECT p.url, r.elapsed_time, r.status
FROM requests r
INNER JOIN pages p ON r.page_id = p.id
WHERE r.created_at >= ? AND status = 200 AND elapsed_time > 2
ORDER BY r.elapsed_time DESC
LIMIT 10;
    `

	return getPages(db, query)
}

func getErrorsPages(db *sql.DB) []PageInfo {
	query := `
SELECT p.url, r.elapsed_time, r.status
FROM requests r
INNER JOIN pages p ON r.page_id = p.id
WHERE r.created_at >= ? AND status != 200
ORDER BY r.elapsed_time DESC
LIMIT 20;
    `

	return getPages(db, query)
}

func getPages(db *sql.DB, query string) []PageInfo {
	startTime := startTime.In(time.UTC)
	startTimeStr := startTime.Format("2006-01-02 15:04:05")

	rows, err := db.Query(query, startTimeStr)
	if err != nil {
		return nil
	}

	defer rows.Close()

	var pages []PageInfo

	for rows.Next() {
		var url string
		var status int
		var elapsedTime float64

		if err := rows.Scan(&url, &elapsedTime, &status); err != nil {
			log.Error("Ошибка получения данных для статистики:", err)
			return nil
		}

		page := PageInfo{
			URL:         url,
			ElapsedTime: elapsedTime,
			Status:      status,
		}

		pages = append(pages, page)
	}

	return pages
}
