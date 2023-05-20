package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

/*
Показывает список доступных тем для теста
*/
func main() {
	var dbFile string
	flag.StringVar(&dbFile, "db", "test.db", "путь к файлу с базой данных SQLITE3")
	flag.Parse()

	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()

	rows, err := db.Query("select id, title from themes")
	if err != nil {
		log.Panic(err)
	}

	for rows.Next() {
		var (
			themeID    int64
			themeTitle string
		)

		if err := rows.Scan(&themeID, &themeTitle); err != nil {
			log.Panic(err)
		}

		fmt.Printf("ID=%d %s\n", themeID, themeTitle)
	}
}
