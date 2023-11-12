package object

import (
	"fmt"
	"time"
)

type author struct {
	name  string
	email string
	now   time.Time
}

type Author interface {
	String() string
}

func NewAuthor(name, email string, now time.Time) *author {
	return &author{name, email, now}
}

func (a *author) String() string {

	t := a.now.Format(time.RFC822Z)
	offset := t[len(t)-5:]
	return fmt.Sprintf("%s <%s> %d %s", a.name, a.email, a.now.Unix(), offset)
}
