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
	Username() string
	Email() string

	Out(string) error
	OutError(error) error
}

type gotContext struct {
	c.Context
	workspaceRoot string
	gotRoot       string
	username      string
	email         string
	w             io.Writer
	e             io.Writer
}

func NewContext(ctx c.Context, workspaceRoot, gotroot, username, email string, out io.Writer, err io.Writer) GotContext {
	return &gotContext{ctx, workspaceRoot, filepath.Join(workspaceRoot, gotroot), username, email, out, err}
}

func (g *gotContext) WorkspaceRoot() string {
	return g.workspaceRoot
}

func (g *gotContext) GotRoot() string {
	return g.gotRoot
}

func (g *gotContext) Username() string {
	return g.username
}

func (g *gotContext) Email() string {
	return g.email
}

func (g *gotContext) Out(msg string) error {

	_, err := g.w.Write([]byte(msg))

	return err
}

func (g *gotContext) OutError(e error) error {

	_, err := g.w.Write([]byte(e.Error()))

	return err
}
