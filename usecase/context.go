package usecase

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fatih/color"
)

type GotContext interface {
	GotContextReader
	GotContextWriter
	GotContextCloser
}

type GotContextReader interface {
	context.Context
	WorkspaceRoot() string
	GotRoot() string
	Username() string
	Email() string
}

type GotContextWriter interface {
	Out(string, ColorAttribute) error
	OutError(error) error
}

type GotContextReaderWriter interface {
	GotContextReader
	GotContextWriter
}

type GotContextCloser interface {
	Close() error
}

type gotContext struct {
	context.Context
	workspaceRoot string
	gotRoot       string
	username      string
	email         string
	w             io.Writer
	e             io.Writer
}

func NewContext(ctx context.Context, workspaceRoot, gotroot, username, email string, out io.Writer, errOut io.Writer) GotContext {

	return &gotContext{ctx, workspaceRoot, filepath.Join(workspaceRoot, gotroot), username, email, out, errOut}
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

	_, err = g.w.Write([]byte(colorize(msg, c)))
	return
}

func colorize(msg string, c ColorAttribute) string {

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

	return color.New(attrs...).Sprint(msg)
}

func (g *gotContext) OutError(e error) error {

	_, err := g.w.Write([]byte(e.Error()))

	return err
}

func (g *gotContext) Close() error {
	return nil
}

type gotContextPager struct {
	*gotContext
	cmd *exec.Cmd
	out io.WriteCloser
}

func NewContextPager(ctx context.Context, workspaceRoot, gotroot, username, email string, out io.Writer, errOut io.Writer) (GotContext, error) {

	pager := "less"
	if p := os.Getenv("GIT_PAGER"); p != "" {
		pager = p
	}
	if p := os.Getenv("PAGER"); p != "" {
		pager = p
	}

	cmd := exec.Command(pager)
	cmd.Env = append(os.Environ(), "LESS=FRX", "LV=-c")

	stdout, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	return &gotContextPager{&gotContext{ctx, workspaceRoot, filepath.Join(workspaceRoot, gotroot), username, email, out, errOut}, cmd, stdout}, nil
}

func (g *gotContextPager) Out(msg string, c ColorAttribute) (err error) {

	_, err = fmt.Fprint(g.out, colorize(msg, c))
	return
}

func (g *gotContextPager) Close() error {

	g.out.Close()
	return g.cmd.Wait()
}
