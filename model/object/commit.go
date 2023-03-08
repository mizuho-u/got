package object

type commit struct {
	*object
}

func NewCommit(parent, tree, author, message string) (*commit, error) {

	content := []byte{}

	content = append(content, []byte("tree "+tree+"\n")...)
	if parent != "" {
		content = append(content, []byte("parent "+parent+"\n")...)
	}
	content = append(content, []byte("author "+author+"\n")...)
	content = append(content, []byte("committer "+author+"\n")...)
	content = append(content, []byte("\n")...)
	content = append(content, []byte(message)...)

	object, err := newObject(content, classCommit)
	if err != nil {
		return nil, err
	}

	return &commit{object: object}, nil
}
