package yaml

import (
	"sort"

	"github.com/goccy/go-yaml/ast"
)

func Suggestions(f *ast.File) []string {
	suggestionsMap := make(map[string]any)
	for _, doc := range f.Docs {
		for _, suggestion := range nodeSuggestion(doc) {
			suggestionsMap[suggestion] = nil
		}
	}
	suggestions := make([]string, 0, len(suggestionsMap))
	for suggestion := range suggestionsMap {
		suggestions = append(suggestions, suggestion)
	}
	sort.Strings(suggestions)
	return suggestions
}

func nodeSuggestion(n ast.Node) []string {
	suggestions := make([]string, 0)
	suggestions = append(suggestions, nodePathToQueryPath(n.GetPath()))

	switch node := n.(type) {
	case *ast.MappingNode:
		for _, valueNode := range node.Values {
			suggestions = append(suggestions, nodeSuggestion(valueNode)...)
		}
	case *ast.MappingValueNode:
		suggestions = append(suggestions, nodeSuggestion(node.Key)...)
		suggestions = append(suggestions, nodeSuggestion(node.Value)...)
	case *ast.SequenceNode:
		for _, valueNode := range node.Values {
			suggestions = append(suggestions, nodeSuggestion(valueNode)...)
		}
	case *ast.DocumentNode:
		suggestions = append(suggestions, nodeSuggestion(node.Body)...)
	}

	return suggestions
}
