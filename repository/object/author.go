package object

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
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

func authorFromString(s string) (*author, error) {

	re := regexp.MustCompile(`^(.*) ?<(.*)> (\d+) (.+)$`)
	match := re.FindStringSubmatch(s)

	if len(match) != 5 {
		return nil, errors.New("invalid author format")
	}

	unixtime, err := strconv.ParseInt(match[3], 10, 64)
	if err != nil {
		return nil, err
	}

	now := time.Unix(unixtime, 0)

	offset, err := time.Parse("-0700", match[4])
	if err != nil {
		return nil, err
	}

	return &author{
		name:  match[1],
		email: match[2],
		now:   now.In(offset.Location()),
	}, nil

}

func (a *author) String() string {

	t := a.now.Format(time.RFC822Z)
	offset := t[len(t)-5:]
	return fmt.Sprintf("%s <%s> %d %s", a.name, a.email, a.now.Unix(), offset)
}
