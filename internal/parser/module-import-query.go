package parser

import sitter "github.com/smacker/go-tree-sitter"

type ModuleImportPair struct {
	moduleName string
	importName string
}

// These are defined by the order of the captures in the query, if the query is
// changed this will need to be updated.
var _IMPORTS_LIST_IDX = uint32(2)
var _MODULE_NAME_IDX = uint32(3)

func GetImportsByModuleFromFile(
	node *sitter.Node,
	sourceCode []byte,
) (map[string][]string, error) {
	importsQuery, err := LoadModuleImportQuery()
	if err != nil {
		return nil, err
	}
	// Parse source code
	qc := sitter.NewQueryCursor()
	qc.Exec(importsQuery, node)
	importsByModule := make(map[string][]string)
	for {
		m, ok := qc.NextMatch()
		if !ok {
			break
		}
		// Apply predicates filtering
		m = qc.FilterPredicates(m, sourceCode)
		currPair := ModuleImportPair{}
		for _, c := range m.Captures {
			if c.Index == _MODULE_NAME_IDX {
				currPair.moduleName = c.Node.Content(sourceCode)
			} else if c.Index == _IMPORTS_LIST_IDX {
				currPair.importName = c.Node.Content(sourceCode)
			}
		}
		if _, ok = importsByModule[currPair.moduleName]; !ok {
			importsByModule[currPair.moduleName] = []string{}
		}
		importsByModule[currPair.moduleName] = append(
			importsByModule[currPair.moduleName],
			currPair.importName,
		)
	}
	return importsByModule, nil
}
