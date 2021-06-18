package execpool

import (
	"context"
	"io"
	"os/exec"
	"sync"
)

// Command is wrapper around exec.Cmd that implements io.Reader interface
type Command struct {
	cmd    *exec.Cmd
	lock   sync.RWMutex
	err    error
	stdin  io.WriteCloser
	stdout io.ReadCloser
	cancel func()
	ctx    context.Context
}

// Read implements io.Reader
func (c *Command) Read(p []byte) (int, error) {
	n, err := c.stdout.Read(p)
	if err == io.EOF {
		if waitErr := c.cmd.Wait(); waitErr != nil {
			c.stop(waitErr)
			return n, waitErr
		}

		c.stop(nil)
		return n, io.EOF
	}

	if err := c.error(); err != nil {
		return n, err
	}

	if err != nil {
		c.stop(err)
	}

	return n, err
}

// stop stops command execution by closing stdout and cancelling
// command context
func (c *Command) stop(err error) {
	c.lock.Lock()
	if c.err == nil {
		c.err = err
	}
	c.lock.Unlock()

	c.stdout.Close()
	c.cancel()
}

// error returns command error
func (c *Command) error() error {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.err
}
