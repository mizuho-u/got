package repository

import (
	"io"
)

type repository struct {
	index            *index
	object           ObjectLoader
	workspace        map[string]WorkspaceEntry
	head             map[string]TreeEntry
	changed          map[string]status
	indexChanges     map[string]status
	workspaceChanges map[string]status
	untracked        []string
}

type WorkspaceOption func(*repository) error

func WithIndex(data io.Reader) WorkspaceOption {

	return func(w *repository) error {

		index, err := NewIndex(IndexSource(data))
		if err != nil {
			return nil
		}

		w.index = index
		return nil
	}

}

func WithObjectLoader(l ObjectLoader) WorkspaceOption {

	return func(r *repository) error {
		r.object = l
		return nil
	}

}

func NewRepository(options ...WorkspaceOption) (*repository, error) {

	index, err := NewIndex()
	if err != nil {
		return nil, err
	}

	ws := &repository{
		index:            index,
		changed:          map[string]status{},
		indexChanges:     map[string]status{},
		workspaceChanges: map[string]status{},
		head:             map[string]TreeEntry{},
		untracked:        []string{},
		workspace:        map[string]WorkspaceEntry{}}

	for _, opt := range options {
		if err := opt(ws); err != nil {
			return nil, err
		}
	}

	return ws, nil

}

func (repo *repository) Index() Index {
	return repo.index
}
