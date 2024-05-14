package pathresolver

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type TsPathResolver struct {
	paths       map[string][]string
	projectRoot string
}

type CompilerOptions struct {
	Paths map[string][]string `json:"paths"`
}

type TsConfig struct {
	CompilerOptions CompilerOptions `json:"compilerOptions"`
}

func removeCommentLinesFromJson(tsConfigFileContents []byte) []byte {
	var tsConfigFileContentsWithoutComments []byte
	inMultiLineComment := false
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
			tsConfigFileContentsWithoutComments = append(tsConfigFileContentsWithoutComments, line...)
		}
	}
	return tsConfigFileContentsWithoutComments
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
	tsConfigFileContents, err := os.ReadFile(filepath.Join(projectRoot, "tsconfig.json"))
	if err != nil {
		return nil, fmt.Errorf("could not read tsconfig file: %w", err)
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
	}, nil
}

func (t *TsPathResolver) ResolveImportPath(importPath string) string {
	if !strings.HasSuffix(importPath, ".ts") {
		importPath = importPath + ".ts"
	}

	for alias, paths := range t.paths {
		aliasPattern := "^" + strings.ReplaceAll(alias, "*", "(.*)") + "$"
		regexpAlias := regexp.MustCompile(aliasPattern)
		if regexpAlias.MatchString(importPath) {
			submatches := regexpAlias.FindStringSubmatch(importPath)
			for _, path := range paths {
				resolvedPath := path
				if strings.Contains(path, "*") && len(submatches) > 1 {
					resolvedPath = strings.Replace(path, "*", submatches[1], 1)
				}
				absolutePath := filepath.Join(t.projectRoot, resolvedPath)
				return absolutePath
			}
		}
	}
	return filepath.Join(t.projectRoot, importPath)
}
