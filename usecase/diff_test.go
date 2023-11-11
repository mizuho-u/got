package usecase_test

import (
	"bytes"
	"path/filepath"
	"testing"
	"time"

	"github.com/mizuho-u/got/usecase"
)

func TestDiff(t *testing.T) {

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
@@ -1,1 +1,1 @@
-one
+hogehoge
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
@@ -1,1 +1,1 @@
-one
+hogehoge
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
@@ -1,1 +0,0 @@
-one
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

			commit(t, dir, "", "", "commit massage", time.Unix(1677142145, 0))

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
			if err := usecase.Diff(newContext(dir, "", "", out, out), false); err != nil {
				t.Error(err)
			}

			if out.String() != tc.expect {
				t.Errorf("expect \n%s, got \n%s", tc.expect, out)
			}

		})

	}

}

func TestDiffCached(t *testing.T) {

	testt := []struct {
		description              string
		newAddedFiles            map[string]string
		newModifiedFiles         []string
		newModifiedContentsFiles map[string]string
		deleted                  []string
		expect                   string
	}{
		{
			description:   "add a empty file",
			newAddedFiles: map[string]string{"a/4.txt": ""},
			expect: `diff --git a/a/4.txt b/a/4.txt
new file mode 100644
index 0000000..e69de29
--- /dev/null
+++ b/a/4.txt
`,
		},
		{
			description:   "add files",
			newAddedFiles: map[string]string{"a/4.txt": "four"},
			expect: `diff --git a/a/4.txt b/a/4.txt
new file mode 100644
index 0000000..ea1f343
--- /dev/null
+++ b/a/4.txt
@@ -0,0 +1,1 @@
+four
`,
		},
		{
			description:              "modify files",
			newModifiedContentsFiles: map[string]string{"a/b/3.txt": "changed"},
			expect: `diff --git a/a/b/3.txt b/a/b/3.txt
index 1d19714..21fb1ec 100644
--- a/a/b/3.txt
+++ b/a/b/3.txt
@@ -1,1 +1,1 @@
-three
+changed
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
@@ -1,1 +0,0 @@
-one
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

			commit(t, dir, "", "", "commit massage", time.Unix(1677142145, 0))

			for file, contents := range tc.newAddedFiles {
				f := createFile(t, dir, file, []byte(contents))
				add(t, dir, f)
			}

			for _, file := range tc.newModifiedFiles {
				modifyFileMode(t, dir, file, 0755)
				add(t, dir, filepath.Join(dir, file))
			}

			for file, contents := range tc.newModifiedContentsFiles {
				f := createFile(t, dir, file, []byte(contents))
				add(t, dir, f)
			}

			for _, path := range tc.deleted {
				removeAll(t, dir, path)
				removeAll(t, dir, ".git/index")
				add(t, dir, dir)
			}

			out := &bytes.Buffer{}
			if err := usecase.Diff(newContext(dir, "", "", out, out), true); err != nil {
				t.Error(err)
			}

			if out.String() != tc.expect {
				t.Errorf("expect \n%s, got \n%s", tc.expect, out)
			}

		})

	}

}
