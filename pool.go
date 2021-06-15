package execpool

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"
)

// Pool of processes
type Pool struct {
	cmd      *exec.Cmd
	lock     *sync.RWMutex
	err      *errReader
	commands chan *Command
}

// New creates a new pool by spinning processCount copies of cmd
func New(cmd *exec.Cmd, processCount int) (*Pool, error) {
	// Fail fast here and don't let to create a pool if the command can't
	// be executed
	if _, err := exec.LookPath(cmd.Path); err != nil {
		return nil, fmt.Errorf("failed to look path: %w", err)
	}

	pool := Pool{
		cmd:      cmd,
		commands: make(chan *Command, processCount),
		lock:     &sync.RWMutex{},
	}

	for i := 0; i < processCount; i++ {
		command, err := pool.newCommand()
		if err != nil {
			return nil, fmt.Errorf("failed to create a new command %d: %w", i, err)
		}

		pool.commands <- command
	}

	return &pool, nil
}

// Exec fetches a waiting process from the pool and attaches stdin
// to this process.
func (p *Pool) Exec(stdin io.Reader) (stdout io.Reader) {
	command := <-p.commands

	if command == nil {
		return p.error()
	}

	go func() {
		newCmd, err := p.newCommand()
		if err != nil {
			p.lock.Lock()
			if p.err == nil {
				p.err = &errReader{fmt.Errorf("failed to add the command in the pool: %w", err)}
				close(p.commands)
			}
			p.lock.Unlock()
			return
		}

		p.commands <- newCmd
	}()

	go func() {
		defer func() {
			if err := command.stdin.Close(); err != nil {
				command.stop(fmt.Errorf("failed to close stdin pipe: %w", err))
			}
		}()

		if _, err := io.Copy(command.stdin, stdin); err != nil {
			command.stop(fmt.Errorf("failed to copy stdin: %w", err))
			return
		}
	}()

	return command
}

// ExecContext takes a waiting process from the pool and attaches stdin to it
// and returns a reader representing its stdout. When ctx id done the pool closes
// stdout and terminate the process.
func (p *Pool) ExecContext(ctx context.Context, stdin io.Reader) (stdout io.Reader) {
	command := p.Exec(stdin).(*Command)

	go func() {
		select {
		case <-ctx.Done():
			command.cancel()
		case <-command.ctx.Done():
		}
	}()

	return command
}

func (p *Pool) error() *errReader {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.err
}

type errReader struct {
	error
}

// Read implements io.Reader
func (e *errReader) Read([]byte) (int, error) {
	return 0, e.error
}
