package usecase

import (
	c "context"
	"io"
	"path/filepath"

	"github.com/fatih/color"
)

type GotContext interface {
	c.Context
	WorkspaceRoot() string
	GotRoot() string
	Username() string
	Email() string

	Out(string, ColorAttribute) error
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

type ColorlizeWriter interface {
	Write(p []byte, color ColorAttribute) (n int, err error)
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

type ColorAttribute uint

const (
	none ColorAttribute = 0
	bold                = 1 << iota
	red
	green
	cyan
)

func (g *gotContext) Out(msg string, c ColorAttribute) (err error) {

	attrs := []color.Attribute{}
	switch {
	case c&bold != 0:
		attrs = append(attrs, color.Bold)
	case c&red != 0:
		attrs = append(attrs, color.FgRed)
	case c&green != 0:
		attrs = append(attrs, color.FgGreen)
	case c&cyan != 0:
		attrs = append(attrs, color.FgCyan)
	default:
	}

	if _, err = color.New(attrs...).Fprint(g.w, msg); err != nil {
		return err
	}

	return err
}

func (g *gotContext) OutError(e error) error {

	_, err := g.w.Write([]byte(e.Error()))

	return err
}
