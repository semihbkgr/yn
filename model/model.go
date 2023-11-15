package model

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbletea"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/lexer"
	"github.com/goccy/go-yaml/parser"

	"github.com/semihbkgr/yn/yaml"
)

type model struct {
	viewport viewport.Model
	input    textinput.Model

	keymap KeyMap

	opts Options

	file *ast.File
}

func initModel() model {
	v := viewport.New(0, 0)
	i := textinput.New()

	return model{
		viewport: v,
		input:    i,
		keymap:   DefaultKeyMap(),
	}
}

func NewModel(opts Options) (tea.Model, error) {
	m := initModel()
	m.opts = opts

	file, err := parser.ParseBytes(opts.Input, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	// fix the tokens' positions
	file, err = parser.Parse(lexer.Tokenize(file.String()), parser.ParseComments)
	if err != nil {
		return nil, err
	}

	m.file = file
	m.viewport.SetContent(yaml.Print(file, ""))
	return m, nil
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, m.keymap.Quit) {
			return m, tea.Quit
		}
		if key.Matches(msg, m.keymap.Up, m.keymap.Down) {
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}
		if key.Matches(msg, m.keymap.Navigate) {
			m.Navigate()
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 1
	}

	var cmds []tea.Cmd

	if !m.input.Focused() {
		cmds = append(cmds, m.input.Focus())
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Sequence(cmds...)
}

func (m model) View() string {
	return fmt.Sprintf("%s\n%s", m.viewport.View(), m.input.View())
}

func (m *model) Navigate() {
	path := m.input.Value()
	m.viewport.SetContent(yaml.Print(m.file, path))
}
