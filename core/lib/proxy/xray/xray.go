package xray

import (
	"bushuray-core/lib"
	"context"
	"fmt"
	"os/exec"
	"path"
	"sync"
)

type XrayCore struct {
	mu      sync.Mutex
	cmd     *exec.Cmd
	cancel  context.CancelFunc
	running bool
	Exited  chan error
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
		x.running = false
		x.mu.Unlock()
		x.Exited <- err
	}()

	return nil
}

func (x *XrayCore) Stop() {
	x.mu.Lock()
	defer x.mu.Unlock()

	if !x.running {
		return
	}
	x.cancel()
	x.running = false
}

func (x *XrayCore) IsRunning() bool {
	x.mu.Lock()
	defer x.mu.Unlock()
	return x.running
}
