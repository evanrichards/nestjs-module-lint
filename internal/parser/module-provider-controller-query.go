package parser

import sitter "github.com/smacker/go-tree-sitter"

// These are defined by the order of the captures in the query, if the query is
// changed this will need to be updated.
var _PROVIDER_CONTROLLER_QUERY_PROVIDER_CONTROLLER_LIST_IDX = uint32(2)
var _PROVIDER_CONTROLLER_QUERY_MODULE_NAME_IDX = uint32(3)

func GetProviderControllersByModuleFromFile(
	node *sitter.Node,
	sourceCode []byte,
) (map[string][]string, error) {
	providerControllersQuery, err := LoadModuleProviderControllerQuery()
	if err != nil {
		return nil, err
	}
	// Parse source code
	qc := sitter.NewQueryCursor()
	qc.Exec(providerControllersQuery, node)
	providerControllersByModule := make(map[string][]string)
	for {
		m, ok := qc.NextMatch()
		if !ok {
			break
		}
		// Apply predicates filtering
		m = qc.FilterPredicates(m, sourceCode)
		currPair := struct {
			moduleName             string
			providerControllerName string
		}{}
		for _, c := range m.Captures {
			if c.Index == _PROVIDER_CONTROLLER_QUERY_MODULE_NAME_IDX {
				currPair.moduleName = c.Node.Content(sourceCode)
			} else if c.Index == _PROVIDER_CONTROLLER_QUERY_PROVIDER_CONTROLLER_LIST_IDX {
				currPair.providerControllerName = c.Node.Content(sourceCode)
			}
		}
		if currPair.providerControllerName == "" || currPair.moduleName == "" {
			continue
		}
		if _, ok = providerControllersByModule[currPair.moduleName]; !ok {
			providerControllersByModule[currPair.moduleName] = []string{}
		}
		providerControllersByModule[currPair.moduleName] = append(
			providerControllersByModule[currPair.moduleName],
			currPair.providerControllerName,
		)
	}
	return providerControllersByModule, nil
}
