package xray

import (
	"bushuray-core/lib"
	"context"
	"fmt"
	"log"
	"os/exec"
	"path"
	"sync"
)

type XrayCore struct {
	mu             sync.Mutex
	cmd            *exec.Cmd
	cancel         context.CancelFunc
	running        bool
	channel_closed bool
	Exited         chan error
}

func (x *XrayCore) Start(stdinPipe []byte) error {
	x.mu.Lock()
	defer x.mu.Unlock()

	if x.running {
		return fmt.Errorf("command is already running")
	}

	ctx, cancel := context.WithCancel(context.Background())

	xraybin := path.Join(lib.GetWorkingDir(), "bin", "xray")

	cmd := exec.CommandContext(ctx, xraybin, "run")
	stdin, err := cmd.StdinPipe()

	if err != nil {
		cancel()
		return fmt.Errorf("failed to get stdin %w", err)
	}

	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		cancel()
		return err
	}

	go func() {
		defer stdin.Close()
		_, _ = stdin.Write(stdinPipe)
	}()

	x.cmd = cmd
	x.cancel = cancel
	x.running = true

	go func() {
		err := cmd.Wait()
		x.mu.Lock()
		defer x.mu.Unlock()
		if ctx.Err() == nil {
			select {
			case x.Exited <- err:
				if x.running {
					close(x.Exited)
				}
				x.running = false
			default:
				// Channel is full or no reader, don't block
			}
		}

	}()

	return nil
}

func (x *XrayCore) Stop() {
	x.mu.Lock()
	defer x.mu.Unlock()
	if !x.channel_closed {
		close(x.Exited)
		x.channel_closed = true
	}
	if x.cmd != nil && x.cmd.Process != nil {
		err := x.cmd.Process.Kill()
		if err != nil {
			log.Println("error killing proces", err)
		}
	}
	x.cancel()
	x.running = false
}

func (x *XrayCore) IsRunning() bool {
	x.mu.Lock()
	defer x.mu.Unlock()
	return x.running
}
