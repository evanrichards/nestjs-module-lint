package main

import (
	"github.com/loop-payments/nestjs-module-lint/internal/parser"
	"log"
	"os"
)

func main() {

	if len(os.Args) < 2 {
		log.Fatal("Usage: go run parser.go <path-to-typescript-file>")
	}
	filePath := os.Args[1]
	sourceCode, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Could not read file: %v", err)
	}
	err = parser.ParseAll(sourceCode)
	if err != nil {
		log.Fatalf("Failed to parse: %v", err)
	}
}
