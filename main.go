package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"sync"
	"time"
)

type Category struct {
	ID       int        `json:"id"`
	Slug     string     `json:"slug"`
	Children []Category `json:"children"`
}

type Pagination struct {
	CurrentPage int    `json:"currentPage"`
	LastPage    int    `json:"lastPage"`
	Path        string `json:"path"`
	PerPage     int    `json:"perPage"`
	Total       int    `json:"total"`
}

type CategoryInfo struct {
	Pagination Pagination `json:"pagination"`
}

var log = logrus.New()

func main() {
	setLogParam()

	db, err := initializeDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	readConfig()

	apiURL := "https://mageo.ru/api/v1/menu"

	response, err := http.Get(apiURL)
	if err != nil {
		fmt.Println("Ошибка при запросе API:", err)
		return
	}
	defer response.Body.Close()

	var rootCategories []Category
	if err := json.NewDecoder(response.Body).Decode(&rootCategories); err != nil {
		fmt.Println("Ошибка при декодировании JSON:", err)
		return
	}

	//resultChan := make(chan Category)
	var wg sync.WaitGroup

	for _, rootCategory := range rootCategories {
		wg.Add(1)
		go processCategory(db, rootCategory, &wg)
	}

	wg.Wait()

	//close(resultChan)
}

func processCategory(db *sql.DB, category Category, wg *sync.WaitGroup) {
	defer wg.Done()

	//fmt.Printf("ID: %d, Slug: %s\n", category.ID, category.Slug)

	apiURL := "https://mageo.ru/api/v1/category/" + category.Slug

	start := time.Now()

	response, err := http.Get(apiURL)
	if err != nil {
		fmt.Println("Ошибка при запросе API:", err)
		return
	}
	defer response.Body.Close()

	elapsedTime := time.Since(start)

	var categoryInfo CategoryInfo
	if err := json.NewDecoder(response.Body).Decode(&categoryInfo); err != nil {
		fmt.Println("Ошибка при декодировании JSON:", err)
		return
	}

	saveDB(db, apiURL, response.StatusCode, elapsedTime)

	//if category.Pages > 5 {
	//	for page := 2; page <= category.Pages; page++ {
	//		pageURL := fmt.Sprintf("%s?page=%d", apiURL, page)
	//		response, err := http.Get(pageURL)
	//		if err != nil {
	//			fmt.Println("Ошибка при запросе страницы:", err)
	//			continue
	//		}
	//		defer response.Body.Close()
	//	}
	//}

	for _, childCategory := range category.Children {
		wg.Add(1)
		go processCategory(db, childCategory, wg)
	}
}

func setLogParam() {
	log.Out = os.Stdout

	file, err := os.OpenFile("./logs/logrus.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.Out = file
	} else {
		log.Info("Failed to log to file, using default stderr")
	}
}
