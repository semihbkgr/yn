package model

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbletea"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"

	"github.com/semihbkgr/yn/yaml"
)

type model struct {
	viewport viewport.Model

	opts Options

	file *ast.File
}

func initModel() model {
	v := viewport.New(0, 0)

	return model{
		viewport: v,
	}
}

func NewModel(opts Options) (tea.Model, error) {
	m := initModel()
	m.opts = opts
	file, err := parser.ParseBytes(opts.Input, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	m.file = file
	m.viewport.SetContent(yaml.Print(file, nil))
	return m, nil
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))) {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return m.viewport.View()
}
