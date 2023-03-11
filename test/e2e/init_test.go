package e2e

import (
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"
)

var bin = flag.String("build", "", "the build path")

func TestInit(t *testing.T) {

	buildPathAbs, err := filepath.Abs(*bin)
	if err != nil {
		t.Fatal("invalid build path")
	}

	tt := []struct {
		description string
		path        func(t testing.TB) string
		expect      string
	}{
		{
			description: "no args",
			path: func(t testing.TB) string {
				t.Cleanup(func() {
					os.RemoveAll(".got")
				})
				return ""
			},
			expect: `Initialized empty Jit repository in .+`,
		},
		{
			description: "provide a abs path",
			path: func(t testing.TB) string {
				return t.TempDir()
			},
			expect: `Initialized empty Jit repository in .+`,
		},
	}

	for _, tc := range tt {

		t.Run(tc.description, func(t *testing.T) {

			out, err := exec.Command(buildPathAbs, "init", tc.path(t)).Output()
			if err != nil {
				t.Fatal("exec got command failed ", err)
			}

			if !regexp.MustCompile(tc.expect).MatchString(string(out)) {

				t.Errorf("out msg not match. expect %s got %s", tc.expect, out)

			}

		})

	}

}
