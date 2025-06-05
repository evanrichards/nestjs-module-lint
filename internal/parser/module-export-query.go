package parser

import sitter "github.com/smacker/go-tree-sitter"

// These are defined by the order of the captures in the query, if the query is
// changed this will need to be updated.
const (
	exportListIndex       = uint32(2)
	exportModuleNameIndex = uint32(3)
)

func ParseModuleExports(
	node *sitter.Node,
	sourceCode []byte,
) (map[string][]string, error) {
	exportsQuery, err := LoadModuleExportQuery()
	if err != nil {
		return nil, err
	}
	// Parse source code
	qc := sitter.NewQueryCursor()
	qc.Exec(exportsQuery, node)
	exportsByModule := make(map[string][]string)
	for {
		m, ok := qc.NextMatch()
		if !ok {
			break
		}
		// Apply predicates filtering
		m = qc.FilterPredicates(m, sourceCode)
		currPair := struct {
			moduleName string
			exportName string
		}{}
		for _, c := range m.Captures {
			if c.Index == exportModuleNameIndex {
				currPair.moduleName = c.Node.Content(sourceCode)
			} else if c.Index == exportListIndex {
				currPair.exportName = c.Node.Content(sourceCode)
			}
		}
		if currPair.exportName == "" || currPair.moduleName == "" {
			continue
		}
		if _, ok = exportsByModule[currPair.moduleName]; !ok {
			exportsByModule[currPair.moduleName] = []string{}
		}
		exportsByModule[currPair.moduleName] = append(
			exportsByModule[currPair.moduleName],
			currPair.exportName,
		)
	}
	return exportsByModule, nil
}
