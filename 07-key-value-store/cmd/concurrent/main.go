package main

import (
	"fmt"
	"key-value-store/store"
	"sync"
)

func main() {
	database := store.New()

	var group sync.WaitGroup

	for worker := 0; worker < 10; worker++ {
		group.Add(1)

		go func(worker int) {
			defer group.Done()

			key := fmt.Sprintf("worker-%d", worker)
			value := fmt.Sprintf("value-%d", worker)

			if err := database.Put(key, value); err != nil {
				panic(err)
			}

			stored, found := database.Get(key)
			fmt.Printf("%s: value=%q found=%t\n", key, stored, found)
		}(worker)
	}

	group.Wait()
	fmt.Println("all workers finished")
}
