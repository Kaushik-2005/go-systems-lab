package main

import (
	"fmt"
	"log"
	"write-ahead-log/wal"
)

func main() {
	logFile, err := wal.Open("wal.log")
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()

	payloads := []string{
		"set language Go",
		"set database systems",
		"delete language",
	}

	for _, payload := range payloads {
		record, err := logFile.Append(payload)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf(
			"sequence=%d payload=%q\n",
			record.Sequence,
			record.Payload,
		)
	}
}
