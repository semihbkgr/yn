package model

import (
	"context"
	"errors"
	"io"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

type Options struct {
	Input []byte
}

func NewOptions(command *cobra.Command) (Options, error) {
	b, err := io.ReadAll(command.InOrStdin())
	if err != nil {
		return Options{}, err
	}

	return Options{
		Input: b,
	}, nil
}

func (o Options) Verify() error {
	if len(o.Input) == 0 {
		return errors.New("input cannot be empty")
	}

	return nil
}

func RunProgram(_ context.Context, opts Options) error {
	if err := opts.Verify(); err != nil {
		return err
	}

	m := NewModel(opts)

	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	_, err := p.Run()
	return err
}
