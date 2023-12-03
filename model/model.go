package model

import (
	"fmt"
	"strings"

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

	navigatePath       string
	navigatedNodes     []ast.Node
	navigatedNodeIndex int

	lineNumber bool
}

func initModel() model {
	v := viewport.New(0, 0)
	i := textinput.New()
	i.ShowSuggestions = true

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

func NewModel(opts Options) tea.Model {
	m := initModel()
	m.opts = opts

	// fix the tokens' positions
	file, _ := parser.Parse(lexer.Tokenize(opts.File.String()), parser.ParseComments)
	m.file = file

	m.input.SetSuggestions(yaml.Suggestions(m.file))

	m.Navigate()

	return m
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
			if m.navigatePath == m.input.Value() {
				m.NextNavigatedNode()
				return m, nil
			}

			m.Navigate()
			m.navigatedNodeIndex = -1
			m.NextNavigatedNode()
			return m, nil
		}
		if key.Matches(msg, m.keymap.LineNumber) {
			m.lineNumber = !m.lineNumber
			m.Navigate()
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 2
		m.input.Width = msg.Width - 10
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
	input := m.input.View()
	if len(m.navigatedNodes) > 0 {
		navigateIndex := fmt.Sprintf("%d/%d", m.navigatedNodeIndex+1, len(m.navigatedNodes))
		input += lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(10)).Render(navigateIndex)
	}

	return fmt.Sprintf("%s\n%s\n%s",
		m.viewport.View(),
		input,
		m.help.View(m.keymap))
}

func (m *model) Navigate() {
	path := m.input.Value()
	m.navigatePath = path
	content, navigatedNodes := yaml.Print(m.file, path, m.lineNumber)
	m.viewport.SetContent(content)
	m.navigatedNodes = navigatedNodes
	if len(navigatedNodes) > 0 {
		m.input.TextStyle = m.input.TextStyle.Foreground(lipgloss.ANSIColor(10))
	} else {
		m.input.TextStyle = m.input.TextStyle.Foreground(lipgloss.ANSIColor(9))
	}
}

func (m *model) Output() string {
	if m.navigatePath == "" || len(m.navigatedNodes) == 0 {
		return ""
	}

	var nodes []string
	for _, node := range m.navigatedNodes {
		nodes = append(nodes, node.String())
	}
	nodesStr := strings.Join(nodes, "\n---\n")

	return fmt.Sprintf("%s\n\n%s\n", m.navigatePath, nodesStr)
}

func (m *model) NextNavigatedNode() {
	if len(m.navigatedNodes) == 0 {
		return
	}

	m.navigatedNodeIndex++
	if m.navigatedNodeIndex >= len(m.navigatedNodes) {
		m.navigatedNodeIndex = 0
	}

	navigatedNode := m.navigatedNodes[m.navigatedNodeIndex]
	scrollLine := navigatedNode.GetToken().Position.Line
	scrollYOffset := scrollLine - m.viewport.Height/2
	m.viewport.SetYOffset(scrollYOffset)
}
