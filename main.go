package main

import (
	"fmt"
	"log"
	"os"

	"encoding/json"
	"github.com/loop-payments/nestjs-module-lint/internal/app"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: nestjs-module-lint <path-to-directory-or-file>")
	}
	inputPath := os.Args[1]
	result, err := app.RunForDirRecursively(inputPath)
	if err != nil {
		log.Fatal(err)
	}
	for _, moduleReport := range result {
		data, _ := json.Marshal(moduleReport)
		fmt.Println(string(data))
	}

}
