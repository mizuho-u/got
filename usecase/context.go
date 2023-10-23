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

	Out(string, Color) error
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
	Write(p []byte, color Color) (n int, err error)
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

type Color string

const (
	none  Color = ""
	red   Color = "red"
	green Color = "green"
)

func (g *gotContext) Out(msg string, c Color) (err error) {

	switch c {
	case red:
		_, err = color.New(color.FgRed).Fprint(g.w, msg)
	case green:
		_, err = color.New(color.FgGreen).Fprint(g.w, msg)
	default:
		_, err = g.w.Write([]byte(msg))
	}

	return err
}

func (g *gotContext) OutError(e error) error {

	_, err := g.w.Write([]byte(e.Error()))

	return err
}
