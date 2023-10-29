package usecase_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/mizuho-u/got/usecase"
)

func TestDiffWorkspaceWithIndex(t *testing.T) {

	testt := []struct {
		description              string
		newModifiedFiles         []string
		newModifiedContentsFiles map[string]string
		delete                   []string
		expect                   string
	}{
		{
			description:              "modify contents",
			newModifiedContentsFiles: map[string]string{"1.txt": "hogehoge"},
			expect: `diff --git a/1.txt b/1.txt
index 43dd47e..48f685c 100644
--- a/1.txt
+++ b/1.txt
`,
		},
		{
			description:      "modify mode",
			newModifiedFiles: []string{"1.txt"},
			expect: `diff --git a/1.txt b/1.txt
old mode 100644
new mode 100755
`,
		},
		{
			description:              "modify contents and mode",
			newModifiedContentsFiles: map[string]string{"1.txt": "hogehoge"},
			newModifiedFiles:         []string{"1.txt"},
			expect: `diff --git a/1.txt b/1.txt
old mode 100644
new mode 100755
index 43dd47e..48f685c
--- a/1.txt
+++ b/1.txt
`,
		},
		{
			description: "delete files",
			delete:      []string{"1.txt"},
			expect: `diff --git a/1.txt b/1.txt
deleted file mode 100644
index 43dd47e..0000000
--- a/1.txt
+++ /dev/null
`,
		},
	}

	for _, tc := range testt {

		t.Run(tc.description, func(t *testing.T) {

			now := time.Now()

			dir := initDir(t)

			f1 := createFile(t, dir, "1.txt", []byte("one"))
			modifyFileTime(t, dir, "1.txt", now, now)
			add(t, dir, f1)

			f2 := createFile(t, dir, "a/2.txt", []byte("two"))
			modifyFileTime(t, dir, "a/2.txt", now, now)
			add(t, dir, f2)

			f3 := createFile(t, dir, "a/b/3.txt", []byte("three"))
			modifyFileTime(t, dir, "a/b/3.txt", now, now)
			add(t, dir, f3)

			commit(t, dir, "commit massage", time.Unix(1677142145, 0))

			for _, file := range tc.newModifiedFiles {
				modifyFileMode(t, dir, file, 0755)
			}

			for file, contents := range tc.newModifiedContentsFiles {
				createFile(t, dir, file, []byte(contents))
			}

			for _, name := range tc.delete {
				removeAll(t, dir, name)
			}

			out := &bytes.Buffer{}
			if code := usecase.Diff(newContext(dir, "", "", out, out), false); code != 0 {
				t.Error("expect exit code 0, got ", code)
			}

			if out.String() != tc.expect {
				t.Errorf("expect \n%s, got \n%s", tc.expect, out)
			}

		})

	}

}
