package parser

import sitter "github.com/smacker/go-tree-sitter"

const (
	nameIndex = uint32(0)
	pathIndex = uint32(1)
)

func ParseImportPaths(
	node *sitter.Node,
	sourceCode []byte,
) (map[string]string, error) {
	importPathQuery, err := LoadImportPathQuery()
	if err != nil {
		return nil, err
	}
	qc := sitter.NewQueryCursor()
	qc.Exec(importPathQuery, node)
	importPathsByImportName := make(map[string]string)
	for {
		m, ok := qc.NextMatch()

		if !ok {
			break
		}
		currImport, currPath := "", ""
		for _, c := range m.Captures {
			if c.Index == nameIndex {
				currImport = c.Node.Content(sourceCode)
			}
			if c.Index == pathIndex {
				currPath = c.Node.Content(sourceCode)
			}
		}
		if currImport != "" && currPath != "" {
			importPathsByImportName[currImport] = currPath
		}
	}
	return importPathsByImportName, nil
}
