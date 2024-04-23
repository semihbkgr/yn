package model

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/spf13/cobra"
)

const (
	FlagFile = "file"
)

type Options struct {
	Input []byte
	File  *ast.File

	verified bool
}

func NewOptions(command *cobra.Command) (*Options, error) {
	file, err := command.Flags().GetString(FlagFile)
	if err != nil {
		return nil, err
	}

	var b []byte
	if file != "" {
		b, err = os.ReadFile(file)
		if err != nil {
			return nil, err
		}
	} else {
		b, err = io.ReadAll(command.InOrStdin())
		if err != nil {
			return nil, err
		}
	}

	return &Options{
		Input: b,
	}, nil
}

func (o *Options) Verify() error {
	if o.verified {
		return nil
	}
	o.verified = true

	if len(o.Input) == 0 {
		return errors.New("input cannot be empty")
	}

	file, err := parser.ParseBytes(o.Input, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("error on parsing input: %w", err)
	}
	o.File = file

	return nil
}

func RunProgram(ctx context.Context, opts *Options) (string, error) {
	if err := opts.Verify(); err != nil {
		return "", err
	}

	m := NewModel(*opts)
	p := tea.NewProgram(m,
		tea.WithAltScreen(),
		tea.WithContext(ctx),
	)

	_, err := p.Run()
	if err != nil {
		return "", err
	}

	return "", nil
}
