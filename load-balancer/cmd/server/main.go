package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9000"
	}

	serverID := os.Getenv("SERVER_ID")
	if serverID == "" {
		serverID = "server-1"
	}

	delay := 0 * time.Millisecond
	if delayText := os.Getenv("DELAY_MS"); delayText != "" {
		delayValue, err := strconv.Atoi(delayText)
		if err != nil || delayValue < 0 {
			log.Fatalf("DELAY_MS must be a non-negative integer, got %q", delayText)
		}
		delay = time.Duration(delayValue) * time.Millisecond
	}

	handler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/health" {
			writer.WriteHeader(http.StatusOK)
			return
		}

		time.Sleep(delay)
		fmt.Fprintf(writer, "%s handled %s %s\n", serverID, request.Method, request.URL.Path)
	})

	address := ":" + port
	log.Printf("%s listening on %s", serverID, address)
	log.Fatal(http.ListenAndServe(address, handler))
}
