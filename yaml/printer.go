package yaml

import (
	"fmt"
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

func (p *Printer) Print(file *ast.File, highlight ast.Node) string {
	tokens := lexer.Tokenize(file.String())
	if len(tokens) == 0 {
		return ""
	}

	var texts []string
	for _, t := range tokens {
		lines := strings.Split(t.Origin, "\n")
		renderFunc := p.DefaultProps.RenderFunc(t)

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

var defaultPrinter = newDefaultPrinter()

func Print(file *ast.File, highlight ast.Node) string {
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
