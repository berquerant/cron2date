package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEndToEnd(t *testing.T) {
	e := newExecutor(t)
	defer e.close()

	t.Run("help", func(t *testing.T) {
		assert.Nil(t, run(e.cmd, "-h"))
	})

	const expr = "0 * * * *"

	t.Run("expected lines without start", func(t *testing.T) {
		out, err := output(e.cmd, "-c", "10", expr)
		assert.Nil(t, err)
		assert.Equal(t, 10, strings.Count(out, "\n"), out)
	})

	for _, tc := range []struct {
		title string
		arg   []string
		want  string
	}{
		{
			title: "count 0",
			arg:   []string{"-s", "2023-12-01 10:00:00", "-c", "0", expr},
			want:  ``,
		},
		{
			title: "count 3",
			arg:   []string{"-s", "2023-12-01 10:00:00", "-c", "3", expr},
			want: `2023-12-01 11:00:00
2023-12-01 12:00:00
2023-12-01 13:00:00
`,
		},
		{
			title: "end + 3 hours",
			arg:   []string{"-s", "2023-12-01 10:00:00", "-e", "2023-12-01 13:00:00", expr},
			want: `2023-12-01 11:00:00
2023-12-01 12:00:00
2023-12-01 13:00:00
`,
		},
		{
			title: "duration + 3.5 hours",
			arg:   []string{"-s", "2023-12-01 10:00:00", "-d", "3h30m", expr},
			want: `2023-12-01 11:00:00
2023-12-01 12:00:00
2023-12-01 13:00:00
`,
		},
	} {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			out, err := output(e.cmd, tc.arg...)
			assert.Nil(t, err)
			assert.Equal(t, tc.want, out)
		})
	}
}

func output(name string, arg ...string) (string, error) {
	b, err := exec.Command(name, arg...).Output()
	return string(b), err
}

func run(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Dir = "."
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type executor struct {
	dir string
	cmd string
}

func newExecutor(t *testing.T) *executor {
	t.Helper()
	e := &executor{}
	e.init(t)
	return e
}

func (e *executor) init(t *testing.T) {
	t.Helper()
	dir, err := os.MkdirTemp("", "cron2date")
	if err != nil {
		t.Fatal(err)
	}
	cmd := filepath.Join(dir, "cron2date")
	// build cron2date command
	if err := run("go", "build", "-o", cmd); err != nil {
		t.Fatal(err)
	}
	e.dir = dir
	e.cmd = cmd
}

func (e *executor) close() {
	os.RemoveAll(e.dir)
}
