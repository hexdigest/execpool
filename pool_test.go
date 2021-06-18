package execpool

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("invalid command", func(t *testing.T) {
		cmd := exec.Command("-this-command-does-not-exists-", "none")
		pool, err := New(cmd, 1)
		require.Error(t, err)
		assert.Nil(t, pool)
	})

	t.Run("grep_with_env_and_files", func(t *testing.T) {
		cmd := exec.Command("grep", "none")
		cmd.Env = []string{"LC_ALL=en_US"}

		f, err := os.Open("go.mod")
		require.NoError(t, err)
		defer f.Close()

		cmd.ExtraFiles = []*os.File{f}

		pool, err := New(cmd, 1)
		require.NoError(t, err)
		rc := pool.Exec(strings.NewReader("this makes sense\nthis is nonesense"))
		b, err := ioutil.ReadAll(rc)
		require.NoError(t, err)
		assert.Equal(t, "this is nonesense\n", string(b))
	})

	t.Run("grep_success", func(t *testing.T) {
		cmd := exec.Command("grep", "none")
		pool, err := New(cmd, 1)
		require.NoError(t, err)
		rc := pool.Exec(strings.NewReader("this makes sense\nthis is nonesense"))
		b, err := ioutil.ReadAll(rc)
		require.NoError(t, err)
		assert.Equal(t, "this is nonesense\n", string(b))
	})

	t.Run("grep_invalid_params", func(t *testing.T) {
		cmd := exec.Command("grep", "--invalid-param")
		pool, err := New(cmd, 1)
		require.NoError(t, err)
		rc := pool.Exec(strings.NewReader("this makes sense\nthis is nonesense"))
		b, err := ioutil.ReadAll(rc)
		assert.Len(t, b, 0)
		require.Error(t, err)
		exitError, ok := err.(*exec.ExitError)
		assert.True(t, ok)
		assert.Equal(t, 2, exitError.ExitCode())
	})

	t.Run("error_reading_from_stdin", func(t *testing.T) {
		cmd := exec.Command("grep", "none")
		pool, err := New(cmd, 1)
		require.NoError(t, err)
		rc := pool.Exec(&errReader{io.ErrShortBuffer})
		b, err := ioutil.ReadAll(rc)
		require.Error(t, err)

		unwrapper, ok := err.(interface{ Unwrap() error })
		require.True(t, ok)
		assert.Equal(t, io.ErrShortBuffer, unwrapper.Unwrap())
		assert.Empty(t, b)
	})
}

func TestPool_ExecContext(t *testing.T) {
	t.Run("sleep_10_killed", func(t *testing.T) {
		cmd := exec.Command("sleep", "10")
		pool, err := New(cmd, 1)
		require.NoError(t, err)
		ctx, cancelFunc := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancelFunc()
		rc := pool.ExecContext(ctx, strings.NewReader(""))
		b, err := ioutil.ReadAll(rc)
		assert.Len(t, b, 0)
		require.Error(t, err)
		exitErr, ok := err.(*exec.ExitError)
		require.True(t, ok)
		assert.Equal(t, -1, exitErr.ExitCode())
		assert.Contains(t, strings.ToLower(exitErr.Error()), "killed")
	})
}
