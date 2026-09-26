package main

import (
	"fmt"
	"key-value-store/store"
	"log"
)

func main() {
	database := store.New()

	if err := database.Put("language", "Go"); err != nil {
		log.Fatal(err)
	}

	value, found := database.Get("language")
	fmt.Printf("get language: value=%q found=%t\n", value, found)

	deleted, err := database.Delete("language")
	if err != nil {
		log.Fatal(err)
	}

	_, found = database.Get("language")
	fmt.Printf("delete language: delete=%t found_after_delete=%t\n", deleted, found)
}
