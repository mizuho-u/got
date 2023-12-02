package repository

import "github.com/mizuho-u/got/repository/object"

type Inspector interface {
	TrackableFile(path string, stat WorkspaceFileStat) (bool, error)
	CompareIndexToWorkspace(entry *IndexEntry, stat WorkspaceFile) status
	CompareTreeToIndex(te object.TreeEntry, ie *IndexEntry) status
}

type inspector struct {
	index IndexReader
	ws    Workspace
}

func NewInspector(index IndexReader, ws Workspace) *inspector {
	return &inspector{index, ws}
}

func (i *inspector) CompareIndexToWorkspace(ie *IndexEntry, f WorkspaceFile) status {

	if ie == nil && f == nil {
		return statusNone
	}

	if ie == nil && f != nil {
		return statusFileUntracked
	}

	if ie != nil && f == nil {
		return statusFileDeleted
	}

	if !i.index.match2(f) {
		return statusFileModified
	}

	return statusNone

}

func (i *inspector) CompareTreeToIndex(te object.TreeEntry, ie *IndexEntry) status {

	if te == nil && ie == nil {
		return statusNone
	}

	if te == nil && ie != nil {
		return statusIndexAdded
	}

	if te != nil && ie == nil {
		return statusFileDeleted
	}

	if ie.oid != te.OID() || ie.permission() != te.Permission() {
		return statusFileModified
	}

	return statusNone

}

func (i *inspector) TrackableFile(path string, stat WorkspaceFileStat) (bool, error) {

	if stat == nil {
		return false, nil
	}

	if !stat.IsDir() {
		return !i.index.tracked(path), nil
	}

	entries, err := i.ws.ListDir(path)
	if err != nil {
		return false, err
	}

	for _, e := range entries {

		if e.IsDir() {
			continue
		}

		trackable, err := i.TrackableFile(e.Name(), e)
		if err != nil {
			return false, err
		}

		return trackable, nil
	}

	for _, e := range entries {

		if !e.IsDir() {
			continue
		}

		trackable, err := i.TrackableFile(e.Name(), e)
		if err != nil {
			return false, err
		}

		return trackable, nil
	}

	return false, nil

}
