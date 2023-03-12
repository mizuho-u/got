package e2e

import (
	"os"
	"os/exec"
	"regexp"
	"testing"
)

func TestInit(t *testing.T) {

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

	build := buildpath(t)
	for _, tc := range tt {

		t.Run(tc.description, func(t *testing.T) {

			out, err := exec.Command(build, "init", tc.path(t)).Output()
			if err != nil {
				t.Fatal("exec got command failed ", err)
			}

			if !regexp.MustCompile(tc.expect).MatchString(string(out)) {

				t.Errorf("out msg not match. expect %s got %s", tc.expect, out)

			}

		})

	}

}
