package e2e

import (
	"path/filepath"
	"testing"
	"time"
)

func TestDiff(t *testing.T) {

	build := buildpath(t)

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

			dir := initDir(t, build)

			f1 := createFile(t, dir, "1.txt", []byte("one"))
			modifyFileTime(t, dir, "1.txt", now, now)
			executeCmd(t, build+" -C "+dir+" add "+f1)

			f2 := createFile(t, dir, "a/2.txt", []byte("two"))
			modifyFileTime(t, dir, "a/2.txt", now, now)
			executeCmd(t, build+" -C "+dir+" add "+f2)

			f3 := createFile(t, dir, "a/b/3.txt", []byte("three"))
			modifyFileTime(t, dir, "a/b/3.txt", now, now)
			executeCmd(t, build+" -C "+dir+" add "+f3)

			executeCmd(t, `echo "commit" | `+build+" -C "+dir+" commit")

			for _, file := range tc.newModifiedFiles {
				modifyFileMode(t, dir, file, 0755)
			}

			for file, contents := range tc.newModifiedContentsFiles {
				createFile(t, dir, file, []byte(contents))
			}

			for _, name := range tc.delete {
				removeAll(t, dir, name)
			}

			out := executeCmd(t, ``+build+" -C "+dir+" diff")
			if out != tc.expect {
				t.Errorf("expect \n%s, got \n%s", tc.expect, out)
			}

		})

	}

}

func TestDiffCached(t *testing.T) {

	build := buildpath(t)

	testt := []struct {
		description              string
		newAddedFiles            map[string]string
		newModifiedFiles         []string
		newModifiedContentsFiles map[string]string
		deleted                  []string
		expect                   string
	}{
		{
			description:   "add files",
			newAddedFiles: map[string]string{"a/4.txt": "four"},
			expect: `diff --git a/a/4.txt b/a/4.txt
new file mode 100644
index 0000000..ea1f343
--- /dev/null
+++ b/a/4.txt
`,
		},
		{
			description:              "modify files",
			newModifiedContentsFiles: map[string]string{"a/b/3.txt": "changed"},
			expect: `diff --git b/a/b/3.txt a/a/b/3.txt
index 1d19714..21fb1ec 100644
--- b/a/b/3.txt
+++ a/a/b/3.txt
`,
		},
		{
			description: "delete files",
			deleted:     []string{"1.txt"},
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

			dir := initDir(t, build)

			f1 := createFile(t, dir, "1.txt", []byte("one"))
			modifyFileTime(t, dir, "1.txt", now, now)
			executeCmd(t, build+" -C "+dir+" add "+f1)

			f2 := createFile(t, dir, "a/2.txt", []byte("two"))
			modifyFileTime(t, dir, "a/2.txt", now, now)
			executeCmd(t, build+" -C "+dir+" add "+f2)

			f3 := createFile(t, dir, "a/b/3.txt", []byte("three"))
			modifyFileTime(t, dir, "a/b/3.txt", now, now)
			executeCmd(t, build+" -C "+dir+" add "+f3)

			executeCmd(t, `echo "commit" | `+build+" -C "+dir+" commit")

			for file, contents := range tc.newAddedFiles {
				f := createFile(t, dir, file, []byte(contents))
				executeCmd(t, build+" -C "+dir+" add "+f)
			}

			for _, file := range tc.newModifiedFiles {
				modifyFileMode(t, dir, file, 0755)
				executeCmd(t, build+" -C "+dir+" add "+filepath.Join(dir, file))
			}

			for file, contents := range tc.newModifiedContentsFiles {
				f := createFile(t, dir, file, []byte(contents))
				executeCmd(t, build+" -C "+dir+" add "+f)
			}

			for _, path := range tc.deleted {
				removeAll(t, dir, path)
				removeAll(t, dir, ".git/index")
				executeCmd(t, build+" -C "+dir+" add "+dir)
			}

			out := executeCmd(t, ``+build+" -C "+dir+" diff --cached")
			if out != tc.expect {
				t.Errorf("expect \n%s, got \n%s", tc.expect, out)
			}

		})

	}

}
