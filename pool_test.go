package execpool

import (
	"context"
	"io/ioutil"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("grep_success", func(t *testing.T) {
		cmd := exec.Command("grep", "none")
		pool, err := New(cmd, 1)
		require.NoError(t, err)
		rc := pool.Exec(strings.NewReader("this makes sence\nthis is nonesense"))
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
