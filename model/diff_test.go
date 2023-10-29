package model

import (
	"bytes"
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

				if lines[i] != tc.expect[i] {
					t.Fatalf("line not match %s %s", lines[i], tc.expect[i])
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
	}

	for _, tc := range testt {
		t.Run("", func(t *testing.T) {

			m := newMyers(tc.input[0], tc.input[1])
			diff := m.diff()

			if diff.String() != tc.expect {
				t.Errorf("expect %s got %s", tc.expect, diff)
			}

		})
	}

}
