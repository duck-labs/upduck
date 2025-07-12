package main

import (
	"log"
	"os"

	"github.com/duck-labs/upduck-v2/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
