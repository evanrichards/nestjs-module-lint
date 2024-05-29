package app

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

func FindOwnersFile(dir string) (string, error) {
	for {
		ownersPath := filepath.Join(dir, "OWNERS")
		if info, err := os.Stat(ownersPath); err == nil {
			if info.IsDir() {
				ownersPath = filepath.Join(ownersPath, "OWNERS")
				if _, err := os.Stat(ownersPath); err == nil {
					return ownersPath, nil
				}
			} else {
				return ownersPath, nil
			}
		}
		if dir == "/" {
			break
		}
		dir = filepath.Dir(dir)
	}
	return "", os.ErrNotExist
}

func ParseOwnersFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			return line, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "UNKNOWN", nil
}
