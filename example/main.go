// Package main demonstrates usage of the jsonrepair library.
package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/kaptinlin/jsonrepair"
)

func main() {
	input := "{name: 'John'}"
	if err := run(os.Stdout, input); err != nil {
		log.Fatalf("Failed to repair JSON: %v", err)
	}
}

func run(w io.Writer, input string) error {
	repaired, err := jsonrepair.Repair(input)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(w, repaired)
	return err
}
