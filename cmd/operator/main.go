package main

import (
	"log"

	"github.com/galxe/spotted-network/cmd/operator/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
