package yaml

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/lexer"
	"github.com/goccy/go-yaml/token"
)

type RenderFunc func(string) string

func ColorRenderFunc(color *color.Color) RenderFunc {
	return func(s string) string {
		return color.Sprint(s)
	}
}

func EmptyRenderFunc() RenderFunc {
	return func(s string) string {
		return s
	}
}

type RenderProperties struct {
	MapKey RenderFunc
	Anchor RenderFunc
	Alias  RenderFunc
	Bool   RenderFunc
	String RenderFunc
	Number RenderFunc
}

func (p RenderProperties) RenderFunc(t *token.Token) RenderFunc {
	if t.Indicator == token.BlockStructureIndicator {
		return p.RenderFunc(t.Prev)
	}

	switch t.PreviousType() {
	case token.AnchorType:
		return p.Anchor
	case token.AliasType:
		return p.Alias
	}

	switch t.NextType() {
	case token.MappingValueType:
		return p.MapKey
	}

	switch t.Type {
	case token.BoolType:
		return p.Bool
	case token.AnchorType:
		return p.Anchor
	case token.AliasType:
		return p.Alias
	case token.StringType, token.SingleQuoteType, token.DoubleQuoteType:
		return p.String
	case token.IntegerType, token.FloatType:
		return p.Number
	}

	return EmptyRenderFunc()
}

type Printer struct {
	DefaultProps     RenderProperties
	HighlightedProps RenderProperties
	LineNumberFormat func(n int, all int) string
}

func (p *Printer) Print(file *ast.File, path string, lineNumber bool) (string, []ast.Node) {
	tokens := lexer.Tokenize(file.String())
	if len(tokens) == 0 {
		return "", nil
	}

	queryResult := QueryTokens(file, path)

	var texts []string
	currentLineNumber := tokens[0].Position.Line
	totalLineNumber := tokens[len(tokens)-1].Position.Line
	for _, t := range tokens {
		lines := strings.Split(t.Origin, "\n")

		_, highlighted := queryResult.TokensMap[*t.Position]
		parentNode := queryResult.ParentNodeMap[*t.Position]
		if t.Indicator == token.BlockStructureIndicator {
			_, highlighted = queryResult.TokensMap[*t.Prev.Position]
			parentNode = queryResult.ParentNodeMap[*t.Prev.Position]
		}
		renderFunc := p.RenderFunc(t, highlighted)

		if len(lines) == 1 {
			line := renderFunc(lines[0])
			linePrefix := ""
			if lineNumber {
				linePrefix = p.LineNumberFormat(currentLineNumber, totalLineNumber)
			}

			if len(texts) == 0 {
				texts = append(texts, linePrefix+line)
				currentLineNumber++
			} else {
				text := texts[len(texts)-1]
				texts[len(texts)-1] = text + line
			}
		} else {
			for idx, src := range lines {
				indentNum := 0
				if highlighted && idx > 0 {
					indentNum = parentNode.GetToken().Position.IndentNum
				}

				line := fmt.Sprintf("%s%s", strings.Repeat(" ", indentNum), renderFunc(src[indentNum:]))
				linePrefix := ""
				if lineNumber {
					linePrefix = p.LineNumberFormat(currentLineNumber, totalLineNumber)
				}

				if idx == 0 {
					if len(texts) == 0 {
						texts = append(texts, linePrefix+line)
						currentLineNumber++
					} else {
						text := texts[len(texts)-1]
						texts[len(texts)-1] = text + line
					}
				} else {
					texts = append(texts, fmt.Sprintf("%s%s", linePrefix, line))
					currentLineNumber++
				}
			}
		}
	}

	return strings.Join(texts, "\n"), queryResult.Nodes
}

func (p *Printer) RenderFunc(t *token.Token, highlighted bool) RenderFunc {
	if highlighted {
		return p.HighlightedProps.RenderFunc(t)
	}

	return p.DefaultProps.RenderFunc(t)
}

var defaultPrinter = newDefaultPrinter()

func Print(file *ast.File, path string, lineNumber bool) (string, []ast.Node) {
	return defaultPrinter.Print(file, path, lineNumber)
}

func newDefaultPrinter() Printer {
	return Printer{
		DefaultProps: RenderProperties{
			MapKey: ColorRenderFunc(color.New(color.FgHiCyan)),
			Anchor: ColorRenderFunc(color.New(color.FgHiYellow)),
			Alias:  ColorRenderFunc(color.New(color.FgHiYellow)),
			Bool:   ColorRenderFunc(color.New(color.FgHiMagenta)),
			String: ColorRenderFunc(color.New(color.FgHiGreen)),
			Number: ColorRenderFunc(color.New(color.FgHiMagenta)),
		},
		HighlightedProps: RenderProperties{
			MapKey: ColorRenderFunc(color.New(color.FgCyan, color.BgHiYellow)),
			Anchor: ColorRenderFunc(color.New(color.FgYellow, color.BgHiYellow)),
			Alias:  ColorRenderFunc(color.New(color.FgYellow, color.BgHiYellow)),
			Bool:   ColorRenderFunc(color.New(color.FgMagenta, color.BgHiYellow)),
			String: ColorRenderFunc(color.New(color.FgGreen, color.BgHiYellow)),
			Number: ColorRenderFunc(color.New(color.FgMagenta, color.BgHiYellow)),
		},
		LineNumberFormat: func(n int, all int) string {
			allLen := len(fmt.Sprintf("%d", all))
			numberLen := len(fmt.Sprintf("%d", n))
			padding := strings.Repeat(" ", allLen-numberLen)
			number := color.New(color.FgHiWhite, color.Bold).Sprintf("%d", n)
			separator := color.New(color.FgHiBlack).Sprint("|")
			return fmt.Sprintf("%s%s%s", padding, number, separator)
		},
	}
}

type QueryResult struct {
	Nodes         []ast.Node
	TokensMap     map[token.Position]*token.Token
	ParentNodeMap map[token.Position]ast.Node
}

func QueryTokens(file *ast.File, path string) QueryResult {
	var nodes []ast.Node
	tokensMap := make(map[token.Position]*token.Token)
	parentNodeMap := make(map[token.Position]ast.Node)
	for _, doc := range file.Docs {
		n := FindNode(doc, path)
		if n == nil {
			continue
		}
		nodes = append(nodes, n)

		tokens := TokensFromNode(n)
		for _, t := range tokens {
			tokensMap[*t.Position] = t
			parentNodeMap[*t.Position] = n
		}
	}

	return QueryResult{
		Nodes:         nodes,
		TokensMap:     tokensMap,
		ParentNodeMap: parentNodeMap,
	}
}

func FindNode(n ast.Node, path string) ast.Node {
	if MatchPaths(n, path) {
		return n
	}

	switch node := n.(type) {
	case *ast.MappingNode:
		for _, valueNode := range node.Values {
			if node := FindNode(valueNode, path); node != nil {
				return node
			}
		}
	case *ast.MappingValueNode:
		if node := FindNode(node.Key, path); node != nil {
			return node
		}
		if node := FindNode(node.Value, path); node != nil {
			return node
		}
	case *ast.SequenceNode:
		for _, valueNode := range node.Values {
			if node := FindNode(valueNode, path); node != nil {
				return node
			}
		}
	case *ast.DocumentNode:
		if node := FindNode(node.Body, path); node != nil {
			return node
		}
	}

	return nil
}

func TokensFromNode(n ast.Node) token.Tokens {
	var tokens token.Tokens
	tokens = append(tokens, n.GetToken())

	switch node := n.(type) {
	case *ast.MappingNode:
		for _, valueNode := range node.Values {
			tokens = append(tokens, TokensFromNode(valueNode)...)
		}
	case *ast.MappingValueNode:
		tokens = append(tokens, TokensFromNode(node.Key)...)
		tokens = append(tokens, TokensFromNode(node.Value)...)
	case *ast.SequenceNode:
		for _, valueNode := range node.Values {
			tokens = append(tokens, TokensFromNode(valueNode)...)
		}
	}

	return tokens
}

func MatchPaths(n ast.Node, path string) bool {
	if path == "" {
		return false
	}

	nodePath := n.GetPath()

	if nodePath == "" {
		return false
	}

	switch node := n.(type) {
	case *ast.MappingNode:
		nodePath = TrimPath(node.Path)
	}

	nodePath = nodePath[1:]

	re := regexp.MustCompile(`\[(\d+)\]`)
	nodePath = re.ReplaceAllString(nodePath, ".$1")
	return nodePath == path
}

func TrimPath(path string) string {
	return path[:strings.LastIndex(path, ".")]
}
