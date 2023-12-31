package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEndToEnd(t *testing.T) {
	e := newExecutor(t)
	defer e.close()

	t.Run("help", func(t *testing.T) {
		assert.Nil(t, run(e.cmd, "-h"))
	})

	t.Run("run", func(t *testing.T) {
		out, err := output(e.cmd, "2100-10-02 10:11:00")
		assert.Nil(t, err)
		assert.Equal(t, "11 10 2 10 *\n", out)
	})
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
	dir, err := os.MkdirTemp("", "date2cron")
	if err != nil {
		t.Fatal(err)
	}
	cmd := filepath.Join(dir, "date2cron")
	// build date2cron command
	if err := run("go", "build", "-o", cmd); err != nil {
		t.Fatal(err)
	}
	e.dir = dir
	e.cmd = cmd
}

func (e *executor) close() {
	os.RemoveAll(e.dir)
}
