package usecase_test

import (
	"bytes"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/mizuho-u/got/io/database"
	"github.com/mizuho-u/got/types"
	"github.com/mizuho-u/got/usecase"
)

func TestBranch(t *testing.T) {

	dir := initDir(t)
	f := createFile(t, dir, "hello.txt", []byte("Hello world.\n"))
	add(t, dir, f)
	commit(t, dir, "Mizuho Ueda", "mi_ueda@u-m.dev", "commit\n", time.Unix(1694356071, 0))

	branchName, _ := types.NewBranchName("topic")
	startPoint, _ := types.NewRevision("")

	err := usecase.Branch(newContext(dir, "", "", &bytes.Buffer{}, &bytes.Buffer{}), branchName, startPoint)
	if err != nil {
		t.Fatal(err)
	}

	db := database.NewFSDB(dir, filepath.Join(dir, ".git"))
	head, err := db.Refs().Head()
	if err != nil {
		t.Fatal(err)
	}
	branch, err := db.Refs().Ref("topic")
	if err != nil {
		t.Fatal(err)
	}

	if head.OID() != branch.OID() {
		t.Fatalf("branch pointing at different commit with HEAD. branch %s HEAD %s", branch.OID(), head.OID())
	}

	branchName, _ = types.NewBranchName("topic")
	startPoint, _ = types.NewRevision("")

	err = usecase.Branch(newContext(dir, "", "", &bytes.Buffer{}, &bytes.Buffer{}), branchName, startPoint)
	if err == nil {
		t.Fatal("expect error but got nil")
	}

}

func TestBranchWithStartPoint(t *testing.T) {

	ws := initDir(t)

	f := createFile(t, ws, "hello.txt", []byte("hello"))
	add(t, ws, f)                                               // b6fc4c620b67d95f953a5c1c1230aaab5db5a1b0
	commit(t, ws, "", "", "commit\n", time.Unix(1697289936, 0)) // 9b931331c93bbead9c578ad356827b073d446875

	f2 := createFile(t, ws, "hello2.txt", []byte("hello2"))
	add(t, ws, f2)                                               // 23294b0610492cf55c1c4835216f20d376a287dd
	commit(t, ws, "", "", "commit2\n", time.Unix(1697289936, 0)) // c0fa525207b23dfee1ed12d1ac14d5ef1d406c6b

	f3 := createFile(t, ws, "hello3.txt", []byte("hello3"))
	add(t, ws, f3)                                               // 96803d198baeec4b042d0e489893118ef2356936
	commit(t, ws, "", "", "commit3\n", time.Unix(1697289936, 0)) // c33fcd8ee7432a15b9d8e847bb875dd068c4bd9e

	testt := []struct {
		branch, sp, expect string
	}{
		{branch: "topic1", sp: "", expect: ""},
		{branch: "topic2", sp: "main", expect: ""},
		{branch: "topic3", sp: "main^", expect: ""},
		{branch: "topic4", sp: "main^^", expect: ""},
		{branch: "topic5", sp: "HEAD^", expect: ""},
		{branch: "topic6", sp: "@", expect: ""},
		{branch: "topic7", sp: "main~1", expect: ""},
		{branch: "topic8", sp: "main~2", expect: ""},
		{branch: "topic9", sp: "main^~1", expect: ""},
		{branch: "topic10", sp: "9b9313", expect: ""},
		{branch: "topic11", sp: "c33fcd8e^", expect: ""},
	}

	for _, tc := range testt {

		branch, err := types.NewBranchName(tc.branch)
		if err != nil {
			t.Fatal(err)
		}

		sp, err := types.NewRevision(tc.sp)
		if err != nil {
			t.Fatal(err)
		}

		err = usecase.Branch(newContext(ws, "", "", &bytes.Buffer{}, &bytes.Buffer{}), branch, sp)
		if err != nil {
			t.Fatal(err)
		}

	}

}

func TestBranchWithInvalidStartPoint(t *testing.T) {

	ws := initDir(t)

	for i := 0; i < 20; i++ {

		f := createFile(t, ws, fmt.Sprintf("hello%d.txt", i), []byte(fmt.Sprintf("hello%d", i)))
		add(t, ws, f)
		commit(t, ws, "", "", fmt.Sprintf("commit%d\n", i), time.Unix(1697289936+int64(i), 0))

	}

	testt := []struct {
		sp, expect string
	}{
		{sp: "b04bfec", expect: "object b04bfec0d64fe8d2492f4f05990517801fa5cc3e is a blob, not a commit"},
		{sp: "notexist", expect: "not a valid object name: notexist"},
		{sp: "dd", expect: `short SHA1 dd is ambiguous
hint: dd91746 commit 2023-10-14 - commit18
hint: ddca9c5 commit 2023-10-14 - commit0
`},
	}

	for i, tc := range testt {

		branch, err := types.NewBranchName(fmt.Sprintf("branch%d", i))
		if err != nil {
			t.Fatal(err)
		}

		sp, err := types.NewRevision(tc.sp)
		if err != nil {
			t.Fatal(err)
		}

		err = usecase.Branch(newContext(ws, "", "", &bytes.Buffer{}, &bytes.Buffer{}), branch, sp)
		if err == nil {
			t.Fatal("expect error but got nil")
		}

		if err.Error() != tc.expect {
			t.Fatalf("expect error %s, but got %s", tc.expect, err)
		}

	}

}
