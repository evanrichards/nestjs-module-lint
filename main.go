package main

import (
	"github.com/loop-payments/nestjs-module-lint/internal/app"
	"log"
	"os"
)

func main() {

	if len(os.Args) < 2 {
		log.Fatal("Usage: go run parser.go <path-to-typescript-file>")
	}
	filePath := os.Args[1]
	err := app.Run(filePath)
	if err != nil {
		log.Fatalf("Failed to parse: %v", err)
	}
}
