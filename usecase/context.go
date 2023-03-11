package usecase

import (
	c "context"
	"path/filepath"
)

type GotContext interface {
	c.Context
	WorkspaceRoot() string
	GotRoot() string
}

type gotContext struct {
	c.Context
	workspaceRoot string
	gotRoot       string
}

func NewContext(ctx c.Context, workspaceRoot, gotroot string) GotContext {
	return &gotContext{ctx, workspaceRoot, filepath.Join(workspaceRoot, gotroot)}
}

func (g *gotContext) WorkspaceRoot() string {
	return g.workspaceRoot
}

func (g *gotContext) GotRoot() string {
	return g.gotRoot
}
