package parser

import sitter "github.com/smacker/go-tree-sitter"

var _NAME_IDX = uint32(0)
var _PATH_IDX = uint32(1)

func GetImportPathsByImportNames(
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
			if c.Index == _NAME_IDX {
				currImport = c.Node.Content(sourceCode)
			}
			if c.Index == _PATH_IDX {
				currPath = c.Node.Content(sourceCode)
			}
		}
		if currImport != "" && currPath != "" {
			importPathsByImportName[currImport] = currPath
		}
	}
	return importPathsByImportName, nil
}
