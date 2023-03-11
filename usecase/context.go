package usecase

import (
	c "context"
	"io"
	"path/filepath"
)

type GotContext interface {
	c.Context
	WorkspaceRoot() string
	GotRoot() string
	Out(string) error
	OutError(error) error
}

type gotContext struct {
	c.Context
	workspaceRoot string
	gotRoot       string
	w             io.Writer
	e             io.Writer
}

func NewContext(ctx c.Context, workspaceRoot, gotroot string, out io.Writer, err io.Writer) GotContext {
	return &gotContext{ctx, workspaceRoot, filepath.Join(workspaceRoot, gotroot), out, err}
}

func (g *gotContext) WorkspaceRoot() string {
	return g.workspaceRoot
}

func (g *gotContext) GotRoot() string {
	return g.gotRoot
}

func (g *gotContext) Out(msg string) error {

	_, err := g.w.Write([]byte(msg))

	return err
}

func (g *gotContext) OutError(e error) error {

	_, err := g.w.Write([]byte(e.Error()))

	return err
}
