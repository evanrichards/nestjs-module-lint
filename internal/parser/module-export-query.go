package parser

import sitter "github.com/smacker/go-tree-sitter"

// These are defined by the order of the captures in the query, if the query is
// changed this will need to be updated.
var _EXPORT_QUERY_EXPORT_LIST_IDX = uint32(2)
var _EXPORT_QUERY_MODULE_NAME_IDX = uint32(3)

func GetExportsByModuleFromFile(
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
			if c.Index == _EXPORT_QUERY_MODULE_NAME_IDX {
				currPair.moduleName = c.Node.Content(sourceCode)
			} else if c.Index == _EXPORT_QUERY_EXPORT_LIST_IDX {
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
