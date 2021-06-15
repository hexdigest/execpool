package execpool

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
)

type Command struct {
	cmd    *exec.Cmd
	lock   sync.RWMutex
	err    error
	stdin  io.WriteCloser
	stdout io.ReadCloser
	cancel func()
	ctx    context.Context
}

// newCommand spins a new process that will be waiting for stdin
func (p *Pool) newCommand() (*Command, error) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, p.cmd.Path)
	cmd.Dir = p.cmd.Dir

	if l := len(p.cmd.ExtraFiles); l > 0 {
		cmd.ExtraFiles = make([]*os.File, l)
		copy(p.cmd.ExtraFiles, cmd.ExtraFiles)
	}

	if l := len(p.cmd.Args); l > 0 {
		cmd.Args = make([]string, len(p.cmd.Args))
		copy(cmd.Args, p.cmd.Args)
	}

	if l := len(p.cmd.Env); l > 0 {
		cmd.Env = make([]string, len(p.cmd.Env))
		copy(p.cmd.Env, cmd.Env)
	}

	c := Command{
		cmd:    cmd,
		cancel: cancelFunc,
		ctx:    ctx,
	}

	var err error

	c.stdin, err = cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	c.stdout, err = cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err = c.cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	return &c, nil
}

// Read implements io.Reader
func (c *Command) Read(p []byte) (int, error) {
	if err := c.error(); err != nil {
		return 0, err
	}

	n, err := c.stdout.Read(p)
	if err == io.EOF {
		if waitErr := c.cmd.Wait(); waitErr != nil {
			c.stop(waitErr)
			return n, waitErr
		}

		c.stop(nil)
		return n, io.EOF
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
	c.err = err
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
