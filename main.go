package main

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/loop-payments/nestjs-module-lint/internal/app"
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

	var wg sync.WaitGroup
	errChan := make(chan error, len(files))

	for _, file := range files {
		wg.Add(1)
		go func(file string) {
			defer wg.Done()
			err := app.Run(file)
			if err != nil {
				errChan <- fmt.Errorf("failed to run app for %s: %w", file, err)
				log.Printf("Failed to parse %s: %v", file, err)
			}
		}(file)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(errChan)

	// Check if there were any errors
	for err := range errChan {
		if err != nil {
			log.Printf("Error encountered: %v", err)
		}
	}
}
