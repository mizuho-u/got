package e2e

import (
	"testing"
)

func TestStatus(t *testing.T) {

	build := buildpath(t)

	testt := []struct {
		description string
		files       []string
		dirs        []string
		tracked     []string
		expect      string
	}{
		{
			description: "lists untracked directories, not their contents",
			files:       []string{"file.txt", "dir/another.txt"},
			tracked:     []string{},
			expect:      "?? dir/\n?? file.txt\n",
		},
		{
			description: "lists untracked files inside tracked directories",
			files:       []string{"a/b/inner.txt", "a/outer.txt", "a/b/c/file.txt"},
			tracked:     []string{"a/b/inner.txt"},
			expect:      "?? a/b/c/\n?? a/outer.txt\n",
		},
		{
			description: "does not list empty untracked directories",
			dirs:        []string{"outer"},
			expect:      "",
		},
		{
			description: "lists untracked directories that indirectly contain files",
			files:       []string{"outer/file.txt", "outer/inner/file.txt", "outer2/inner2/innerinnner2/file2.txt"},
			expect:      "?? outer/\n?? outer2/\n",
		},
	}

	for _, tc := range testt {

		t.Run(tc.description, func(t *testing.T) {

			dir := initDir(t, build)

			for _, d := range tc.dirs {
				createDir(t, dir, d)
			}

			fnames := map[string]string{}
			for _, f := range tc.files {
				abs := createFile(t, dir, f, []byte("hello.\n"))
				fnames[f] = abs
			}

			for _, f := range tc.tracked {

				executeCmd(t, build+" -C "+dir+" add "+fnames[f])
				executeCmd(t, `echo "commit" | `+build+" -C "+dir+" commit")

			}

			out := executeCmd(t, build+" -C "+dir+" status")

			if out != tc.expect {
				t.Errorf("expect \n%s, got \n%s", tc.expect, out)
			}

		})

	}

}
