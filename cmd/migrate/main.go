package main

import (
	"database/sql"
	"flag"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

/*
Скрипт создания новой базы данных для telegram бота
*/
func main() {
	var language string

	flag.StringVar(&language, "language", "", "filepath to test db")
	flag.Parse()

	if language == "" {
		log.Panic("db is required argument")
	}

	db, err := sql.Open("sqlite3", language+".db")
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS "themes" (
			"id"	INTEGER NOT NULL UNIQUE,
			"title"	TEXT NOT NULL UNIQUE,
			PRIMARY KEY("id" AUTOINCREMENT)
		)
	`)
	if err != nil {
		log.Panic(err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS "questions" (
			"id"	INTEGER NOT NULL UNIQUE,
			"txt"	TEXT NOT NULL,
			"theme_id"	INTEGER NOT NULL,
			PRIMARY KEY("id" AUTOINCREMENT)
		)
	`)
	if err != nil {
		log.Panic(err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS "answers" (
			"id"	INTEGER NOT NULL UNIQUE,
			"txt"	TEXT NOT NULL,
			"valid"	INTEGER NOT NULL,
			"question_id"	INTEGER NOT NULL,
			PRIMARY KEY("id" AUTOINCREMENT)
		)
	`)
	if err != nil {
		log.Panic(err)
	}

	log.Println("База данных создана")
}
