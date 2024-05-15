package main

import (
	"github.com/loop-payments/nestjs-module-lint/internal/app"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: nestjs-module-lint <path-to-directory-or-file>")
	}
	inputPath := os.Args[1]

	info, err := os.Stat(inputPath)
	if err != nil {
		log.Fatalf("Failed to access path: %v", err)
	}

	var files []string
	if info.IsDir() {
		files, err = app.FindTSFiles(inputPath)
		if err != nil {
			log.Fatalf("Failed to find TypeScript files: %v", err)
		}
	} else {
		files = []string{inputPath}
	}

	for _, file := range files {
		err := app.Run(file)
		if err != nil {
			log.Printf("Failed to parse %s: %v", file, err)
		}
	}
}
