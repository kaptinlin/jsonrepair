// Package main demonstrates usage of the jsonrepair library.
package main

import (
	"fmt"
	"log"

	"github.com/kaptinlin/jsonrepair"
)

func main() {
	// The following is invalid JSON: it consists of JSON contents copied from
	// a JavaScript code base, where the keys are missing double quotes,
	// and strings are using single quotes:
	input := "{name: 'John'}"

	repaired, err := jsonrepair.Repair(input)
	if err != nil {
		log.Fatalf("Failed to repair JSON: %v", err)
	}

	fmt.Println(repaired) // Output: {"name": "John"}
}
