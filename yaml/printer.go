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
}

func (p *Printer) Print(file *ast.File, highlight string) string {
	tokens := lexer.Tokenize(file.String())
	if len(tokens) == 0 {
		return ""
	}

	highlightTokens := QueryTokens(file, highlight)

	var texts []string
	for _, t := range tokens {
		lines := strings.Split(t.Origin, "\n")
		_, highlighted := highlightTokens[*t.Position]
		renderFunc := p.RenderFunc(t, highlighted)

		if len(lines) == 1 {
			line := renderFunc(lines[0])
			if len(texts) == 0 {
				texts = append(texts, line)
			} else {
				text := texts[len(texts)-1]
				texts[len(texts)-1] = text + line
			}
		} else {
			for idx, src := range lines {
				line := renderFunc(src)
				if idx == 0 {
					if len(texts) == 0 {
						texts = append(texts, line)
					} else {
						text := texts[len(texts)-1]
						texts[len(texts)-1] = text + line
					}
				} else {
					texts = append(texts, fmt.Sprintf("%s", line))
				}
			}
		}
	}
	return strings.Join(texts, "\n")
}

func (p *Printer) RenderFunc(t *token.Token, highlighted bool) RenderFunc {
	if highlighted {
		return p.HighlightedProps.RenderFunc(t)
	}

	return p.DefaultProps.RenderFunc(t)
}

var defaultPrinter = newDefaultPrinter()

func Print(file *ast.File, highlight string) string {
	return defaultPrinter.Print(file, highlight)
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
			MapKey: ColorRenderFunc(color.New(color.FgHiCyan, color.BgYellow)),
			Anchor: ColorRenderFunc(color.New(color.FgHiYellow, color.BgYellow)),
			Alias:  ColorRenderFunc(color.New(color.FgHiYellow, color.BgYellow)),
			Bool:   ColorRenderFunc(color.New(color.FgHiMagenta, color.BgYellow)),
			String: ColorRenderFunc(color.New(color.FgHiGreen, color.BgYellow)),
			Number: ColorRenderFunc(color.New(color.FgHiMagenta, color.BgYellow)),
		},
	}
}

func QueryTokens(file *ast.File, path string) map[token.Position]*token.Token {
	//TODO support multi docs yaml
	doc := file.Docs[0]
	n := FindNode(doc, path)
	if n == nil {
		return nil
	}

	tokens := TokensFromNode(n)
	tokenMap := make(map[token.Position]*token.Token)
	for _, t := range tokens {
		tokenMap[*t.Position] = t
	}

	return tokenMap
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
