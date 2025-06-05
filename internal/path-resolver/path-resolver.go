package pathresolver

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

type TsPathResolver struct {
	paths       map[string][]string
	projectRoot string

	// Cache compiled regexes to avoid recompilation
	regexCache map[string]*regexp.Regexp
	regexMutex sync.RWMutex
}

type CompilerOptions struct {
	Paths map[string][]string `json:"paths"`
}

type TsConfig struct {
	CompilerOptions CompilerOptions `json:"compilerOptions"`
}

func removeCommentLinesFromJson(tsConfigFileContents []byte) []byte {
	var builder strings.Builder
	inMultiLineComment := false

	// Pre-allocate approximate capacity to reduce allocations
	builder.Grow(len(tsConfigFileContents))

	for _, line := range strings.Split(string(tsConfigFileContents), "\n") {
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, "/*") {
			inMultiLineComment = true
			continue
		}
		if strings.HasSuffix(trimmedLine, "*/") {
			inMultiLineComment = false
			continue
		}
		if !inMultiLineComment && !strings.HasPrefix(trimmedLine, "//") {
			builder.WriteString(line)
			builder.WriteByte('\n')
		}
	}
	return []byte(builder.String())
}

func ParseTsConfigFile(tsConfigFileContents []byte) (*TsConfig, error) {
	var tsConfig TsConfig
	err := json.Unmarshal(removeCommentLinesFromJson(tsConfigFileContents), &tsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal tsconfig file: %w", err)
	}
	return &tsConfig, nil
}

func NewTsPathResolverFromPath(projectRoot string) (*TsPathResolver, error) {
	tsConfigPath := filepath.Join(projectRoot, "tsconfig.json")
	tsConfigFileContents, err := os.ReadFile(tsConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("tsconfig.json not found in project root (%s). This tool requires a TypeScript project with tsconfig.json", projectRoot)
		}
		return nil, fmt.Errorf("could not read tsconfig.json: %w", err)
	}
	return NewTsPathResolver(tsConfigFileContents, projectRoot)
}

func NewTsPathResolver(tsConfigFileContents []byte, projectRoot string) (*TsPathResolver, error) {
	tsConfig, err := ParseTsConfigFile(tsConfigFileContents)
	if err != nil {
		return nil, err
	}
	return &TsPathResolver{
		paths:       tsConfig.CompilerOptions.Paths,
		projectRoot: projectRoot,
		regexCache:  make(map[string]*regexp.Regexp),
	}, nil
}

func (t *TsPathResolver) ResolveImportPath(importingFileDir, importPath string) string {
	if !strings.HasSuffix(importPath, ".ts") {
		importPath = importPath + ".ts"
	}

	for alias, paths := range t.paths {
		aliasPattern := "^" + strings.ReplaceAll(alias, "*", "(.*)") + "$"

		// Check cache first
		t.regexMutex.RLock()
		regexpAlias, exists := t.regexCache[aliasPattern]
		t.regexMutex.RUnlock()

		if !exists {
			// Compile and cache the regex
			compiledRegex, err := regexp.Compile(aliasPattern)
			if err != nil {
				continue // Skip invalid patterns
			}
			t.regexMutex.Lock()
			t.regexCache[aliasPattern] = compiledRegex
			t.regexMutex.Unlock()
			regexpAlias = compiledRegex
		}

		if regexpAlias.MatchString(importPath) {
			submatches := regexpAlias.FindStringSubmatch(importPath)
			var fallbackPath string
			for i, path := range paths {
				resolvedPath := path
				if strings.Contains(path, "*") && len(submatches) > 1 {
					resolvedPath = strings.Replace(path, "*", submatches[1], 1)
				}
				absolutePath := filepath.Join(t.projectRoot, resolvedPath)
				cleanPath := filepath.Clean(absolutePath)

				// Store first path as fallback
				if i == 0 {
					fallbackPath = cleanPath
				}

				// Check if file exists, if so return it
				if _, err := os.Stat(cleanPath); err == nil {
					return cleanPath
				}
			}

			// If no files exist, return the first path as fallback
			if fallbackPath != "" {
				return fallbackPath
			}
		}
	}
	if strings.HasPrefix(importPath, ".") {
		absolutePath := filepath.Join(importingFileDir, importPath)
		return filepath.Clean(absolutePath)
	}
	return filepath.Join(t.projectRoot, importPath)
}
