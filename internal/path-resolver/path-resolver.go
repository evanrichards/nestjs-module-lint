package pathresolver

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type TsPathResolver struct {
	tsConfigPath string
	paths        map[string][]string
	projectRoot  string
}

type CompilerOptions struct {
	Paths map[string][]string `json:"paths"`
}

type TsConfig struct {
	CompilerOptions CompilerOptions `json:"compilerOptions"`
}

func removeCommentLinesFromJson(tsConfigFileContents []byte) []byte {
	var tsConfigFileContentsWithoutComments []byte
	for _, line := range strings.Split(string(tsConfigFileContents), "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "//") {
			tsConfigFileContentsWithoutComments = append(tsConfigFileContentsWithoutComments, line...)
		}
	}
	return tsConfigFileContentsWithoutComments
}

func ParseTsConfigFile(tsConfigFileContents []byte) (*TsConfig, error) {
	var tsConfig *TsConfig
	err := json.Unmarshal(removeCommentLinesFromJson(
		tsConfigFileContents), &tsConfig)
	if err != nil {
		return nil, err
	}
	return tsConfig, nil

}

func NewTsPathResolverFromPath(tsConfigPath string) (*TsPathResolver, error) {
	tsConfigFileContents, err := os.ReadFile(tsConfigPath)
	if err != nil {
		return nil, err
	}
	return NewTsPathResovler(tsConfigFileContents, tsConfigPath, filepath.Dir(tsConfigPath))
}

func NewTsPathResovler(tsConfigFileContents []byte, tsConfigPath string, projectRoot string) (*TsPathResolver, error) {
	tsConfig, err := ParseTsConfigFile(tsConfigFileContents)
	if err != nil {
		return nil, err
	}
	return &TsPathResolver{
		tsConfigPath: tsConfigPath,
		paths:        tsConfig.CompilerOptions.Paths,
		projectRoot:  projectRoot,
	}, nil
}

func (t *TsPathResolver) ResolveImportPath(importPath string) string {
	for key, value := range t.paths {
		if strings.HasPrefix(importPath, key) {
			return filepath.Join(t.projectRoot, value[0], strings.TrimPrefix(importPath, key))
		}
	}
	return filepath.Join(t.projectRoot, importPath)
}
