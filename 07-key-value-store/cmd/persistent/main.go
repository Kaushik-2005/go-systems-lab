package main

import (
	"fmt"
	"key-value-store/store"
	"log"
)

func main() {
	database, err := store.Open("data.log")
	if err != nil {
		log.Fatal(err)
	}

	if err := database.Put("language", "Go"); err != nil {
		log.Fatal(err)
	}

	if err := database.Put("database", "systems"); err != nil {
		log.Fatal(err)
	}

	deleted, err := database.Delete("language")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("deleted language: %t\n", deleted)

	if err := database.Compact(); err != nil {
		log.Fatal(err)
	}

	if err := database.Close(); err != nil {
		log.Fatal(err)
	}

	recovered, err := store.Open("data.log")
	if err != nil {
		log.Fatal(err)
	}

	defer recovered.Close()

	language, languageFound := recovered.Get("language")
	databaseValue, databaseFound := recovered.Get("database")

	fmt.Printf("recovered language: value=%q found=%t\n", language, languageFound)
	fmt.Printf("recovered database: value=%q found=%t\n", databaseValue, databaseFound)
}
