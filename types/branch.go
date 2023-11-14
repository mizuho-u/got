package types

import (
	"fmt"
	"regexp"
)

type BranchName interface {
	fmt.Stringer
}

type branchName string

var invalidRefName []struct {
	name   string
	reason string
} = []struct {
	name   string
	reason string
}{
	{name: `^\.`, reason: `it begins with "."`},
	{name: `\/\.`, reason: `it has "/."`},
	{name: `\.\.`, reason: `it has double dots ".."`},
	{name: `\/$`, reason: `it ends with a slash "/"`},
	{name: `\.lock$`, reason: `it ends with ".lock"`},
	{name: `@\{`, reason: `it contains a "@{" portion`},
	{name: `[\x00-\x20\x7f]`, reason: `it has ASHCII control characters, tabs or spaces`},
	{name: `[*:?\[\\^~]`, reason: `it has "*", ":", "?", "[", "\", "^", "~"`},
}

func NewBranchName(s string) (branchName, error) {

	for _, invalid := range invalidRefName {

		if regexp.MustCompile(invalid.name).MatchString(s) {
			return "", fmt.Errorf("%s is not a valid branch name. reason %s", s, invalid.reason)
		}

	}

	return branchName(s), nil
}

func (b branchName) String() string {
	return string(b)
}

func validateRefName(s string) error {

	for _, invalid := range invalidRefName {

		if regexp.MustCompile(invalid.name).MatchString(s) {
			return fmt.Errorf("%s is not a valid branch name. reason %s", s, invalid.reason)
		}

	}

	return nil

}
