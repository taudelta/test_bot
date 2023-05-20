package main

import (
	"flag"
	"log"

	"github.com/taudelta/test_bot/internal/importer"
)

func main() {
	var testDescriptionFile string
	var dbOutputFile string
	flag.StringVar(&testDescriptionFile, "test_file", "test.json", "Path to test description file")
	flag.StringVar(&dbOutputFile, "db", "test.db", "Path to test db")
	flag.Parse()

	err := importer.ImportTests(testDescriptionFile, dbOutputFile)
	if err != nil {
		log.Panic(err)
	}
}
