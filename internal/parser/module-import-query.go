package parser

import sitter "github.com/smacker/go-tree-sitter"

// These are defined by the order of the captures in the query, if the query is
// changed this will need to be updated.
const (
	importsListIndex      = uint32(2)
	importModuleNameIndex = uint32(3)
)

func ParseModuleImports(
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
		currPair := struct {
			moduleName string
			importName string
		}{}
		for _, c := range m.Captures {
			if c.Index == importModuleNameIndex {
				currPair.moduleName = c.Node.Content(sourceCode)
			} else if c.Index == importsListIndex {
				currPair.importName = c.Node.Content(sourceCode)
			}
		}
		if currPair.importName == "" || currPair.moduleName == "" {
			continue
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
