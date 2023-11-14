package types

import "testing"

func TestParseRevision(t *testing.T) {

	testt := []struct {
		revision string
		expect   string
	}{
		{revision: "master^", expect: "(parent (ref master))"},
		{revision: "@^", expect: "(parent (ref HEAD))"},
		{revision: "HEAD~42", expect: "(ancestor (ref HEAD) 42)"},
		{revision: "master^^", expect: "(parent (parent (ref master)))"},
		{revision: "abc123~3", expect: "(ancestor (ref abc123) 3)"},
	}

	for _, tc := range testt {

		rev, err := parseRevision(tc.revision)
		if err != nil {
			t.Fatal(err)
		}

		if rev.String() != tc.expect {
			t.Fatalf("expect %s, got %s", tc.expect, rev)
		}

	}

}
