package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEndToEnd(t *testing.T) {
	bin := buildBinary(t)

	t.Run("help", func(t *testing.T) {
		assert.NoError(t, runCmd(bin, "-h"))
	})

	const expr = "0 * * * *"

	t.Run("expected lines without start", func(t *testing.T) {
		out, err := outputCmd(bin, "-c", "10", expr)
		require.NoError(t, err)
		assert.Equal(t, 10, strings.Count(out, "\n"), out)
	})

	t.Run("end without start uses local timezone consistently", func(t *testing.T) {
		jst := time.FixedZone("Asia/Tokyo", 9*3600)
		endStr := time.Now().In(jst).Add(2*time.Hour + 30*time.Minute).Format(time.DateTime)
		cmd := exec.Command(bin, "-e", endStr, expr)
		cmd.Env = append(os.Environ(), "TZ=Asia/Tokyo")
		b, err := cmd.Output()
		require.NoError(t, err)
		lines := strings.Count(string(b), "\n")
		assert.GreaterOrEqual(t, lines, 2, string(b))
		assert.LessOrEqual(t, lines, 3, string(b))
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
		t.Run(tc.title, func(t *testing.T) {
			out, err := outputCmd(bin, tc.arg...)
			require.NoError(t, err)
			assert.Equal(t, tc.want, out)
		})
	}
}

func outputCmd(name string, arg ...string) (string, error) {
	b, err := exec.Command(name, arg...).Output()
	return string(b), err
}

func runCmd(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Dir = "."
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func buildBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "cron2date")
	if err := runCmd("go", "build", "-o", bin); err != nil {
		t.Fatal(err)
	}
	return bin
}
