package repository

import (
	"bytes"
	"strings"
	"testing"
)

func TestInputIntoLines(t *testing.T) {

	testt := []struct {
		input  string
		expect []string
	}{
		{
			input:  "hello\nworld\n",
			expect: []string{"hello", "world"},
		},
		{
			input:  "hello\n",
			expect: []string{"hello"},
		},
		{
			input:  "hello  \n",
			expect: []string{"hello  "},
		},
		{
			input:  "hello",
			expect: []string{"hello"},
		},
		{
			input:  "      \n       \n",
			expect: []string{"      ", "       "},
		},
	}

	for _, tc := range testt {

		t.Run(tc.input, func(t *testing.T) {

			lines, err := lines(bytes.NewBufferString(tc.input))
			if err != nil {
				t.Fatal(err)

			}

			if len(lines) != len(tc.expect) {
				t.Fatalf("len not match %d %d", len(lines), len(tc.expect))
			}

			for i := 0; i < len(lines); i++ {

				if lines[i].text != tc.expect[i] {
					t.Fatalf("line not match %s %s", lines[i].text, tc.expect[i])
				}

			}

		})

	}

}

func TestMyers(t *testing.T) {

	testt := []struct {
		input  [][]string
		expect string
	}{
		{
			input: [][]string{
				{"A", "B", "C", "A", "B", "B", "A"},
				{"C", "B", "A", "B", "A", "C"}},
			expect: `-A
-B
 C
+B
 A
 B
-B
 A
+C
`,
		},
		{
			input: [][]string{
				{""},
				{"C", "B", "A", "B", "A", "C"}},
			expect: `+C
+B
+A
+B
+A
+C
`,
		},
	}

	for _, tc := range testt {
		t.Run("", func(t *testing.T) {

			buf := bytes.NewBufferString(strings.Join(tc.input[0], "\n"))
			a, err := lines(buf)
			if err != nil {
				t.Fatal(err)
			}

			buf = bytes.NewBufferString(strings.Join(tc.input[1], "\n"))
			b, err := lines(buf)
			if err != nil {
				t.Fatal(err)
			}

			m := newMyers(a, b)
			diff := m.diff()

			if diff.String() != tc.expect {
				t.Errorf("expect %s got %s", tc.expect, diff)
			}

		})
	}

}

func TestHunks(t *testing.T) {

	testt := []struct {
		description string
		input       [][]string
		expect      []string
	}{
		{
			description: "test single hunk",
			input: [][]string{
				{"A", "B", "C", "A", "B", "B", "A"},
				{"C", "B", "A", "B", "A", "C"}},
			expect: []string{`@@ -1,7 +1,6 @@
-A
-B
 C
+B
 A
 B
-B
 A
+C
`},
		},
		{
			description: "test a hunk window",
			input: [][]string{
				{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L"},
				{"A", "B", "C", "D", "D", "F", "G", "H", "I", "J", "K", "L"}},
			expect: []string{`@@ -2,7 +2,7 @@
 B
 C
 D
-E
+D
 F
 G
 H
`},
		},
		{
			description: "test multi hunks",
			input: [][]string{
				{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L"},
				{"A", "A", "C", "D", "E", "F", "G", "H", "I", "I", "K", "L"}},
			expect: []string{`@@ -1,5 +1,5 @@
 A
-B
+A
 C
 D
 E
`, `@@ -7,6 +7,6 @@
 G
 H
 I
-J
+I
 K
 L
`},
		},
		{
			description: "test add contents",
			input: [][]string{
				{""},
				{"A", "B", "C", "D"}},
			expect: []string{`@@ -0,0 +1,4 @@
+A
+B
+C
+D
`},
		},
		{
			description: "test delete contents",
			input: [][]string{
				{"A", "B", "C", "D"},
				{""}},
			expect: []string{`@@ -1,4 +0,0 @@
-A
-B
-C
-D
`},
		},
	}

	for _, tc := range testt {
		t.Run(tc.description, func(t *testing.T) {

			buf := bytes.NewBufferString(strings.Join(tc.input[0], "\n"))
			a, err := lines(buf)
			if err != nil {
				t.Fatal(err)
			}

			buf = bytes.NewBufferString(strings.Join(tc.input[1], "\n"))
			b, err := lines(buf)
			if err != nil {
				t.Fatal(err)
			}

			m := newMyers(a, b)
			diff := m.diff()
			hunks := diff.hunks()

			if len(hunks) != len(tc.expect) {
				t.Fatalf("hunk length not match expect %d got %d", len(tc.expect), len(hunks))
			}

			for i, hunk := range hunks {

				if hunk.String() != tc.expect[i] {
					t.Errorf("expect\n%s got\n%s", tc.expect[i], hunk)

				}

			}

		})
	}

}
