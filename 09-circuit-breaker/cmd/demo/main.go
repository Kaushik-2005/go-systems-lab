package main

import (
	"errors"
	"fmt"
	"time"

	"circuit-breaker/breaker"
)

func main() {
	circuit := breaker.New(3, 2*time.Second)

	dependencyCalls := 0

	for request := 1; request <= 5; request++ {
		err := circuit.Execute(func() error {
			dependencyCalls++

			return errors.New("dependency failed")
		})

		fmt.Printf(
			"request=%d error=%v state=%s dependency_calls=%d\n",

			request,
			err,
			circuit.State(),
			dependencyCalls,
		)
	}

	fmt.Println("waiting for reset timeout...")
	time.Sleep(2100 * time.Millisecond)

	err := circuit.Execute(func() error {
		dependencyCalls++
		return errors.New("dependency still failing")
	})

	fmt.Printf(
		"failed recovery probe error=%v state=%s dependency_calls=%d\n",
		err,
		circuit.State(),
		dependencyCalls,
	)

	fmt.Println("waiting for another reset timeout...")
	time.Sleep(2100 * time.Millisecond)

	err = circuit.Execute(func() error {
		dependencyCalls++
		return nil
	})

	fmt.Printf(
		"successful recovery probe error=%v state=%s dependency_calls=%d\n",
		err,
		circuit.State(),
		dependencyCalls,
	)
}
