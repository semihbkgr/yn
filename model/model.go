package model

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/lexer"
	"github.com/goccy/go-yaml/parser"

	"github.com/semihbkgr/yn/yaml"
)

type model struct {
	viewport viewport.Model
	input    textinput.Model
	help     help.Model

	keymap KeyMap

	opts Options

	file *ast.File
}

func initModel() model {
	v := viewport.New(0, 0)
	i := textinput.New()
	i.TextStyle = lipgloss.NewStyle().Bold(true)

	h := help.New()
	h.ShowAll = false
	h.ShortSeparator = "    "

	return model{
		viewport: v,
		input:    i,
		help:     h,
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
	m.Navigate()

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
		m.viewport.Height = msg.Height - 2
		m.input.Width = msg.Width
		m.help.Width = msg.Width
		return m, nil
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
	return fmt.Sprintf("%s\n%s\n%s", m.viewport.View(), m.input.View(), m.help.View(m.keymap))
}

func (m *model) Navigate() {
	path := m.input.Value()
	content, navigated := yaml.Print(m.file, path)
	m.viewport.SetContent(content)
	if navigated {
		m.input.TextStyle = m.input.TextStyle.Foreground(lipgloss.ANSIColor(10))
	} else {
		m.input.TextStyle = m.input.TextStyle.Foreground(lipgloss.ANSIColor(9))
	}
}
