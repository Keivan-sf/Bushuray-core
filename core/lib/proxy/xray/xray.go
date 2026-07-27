package xray

import (
	"bushuray-core/utils"
	"context"
	"fmt"
	"log"
	"os/exec"
	"os/user"
	"strconv"
	"sync"
	"syscall"
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
	return x.startWithCredential(stdinPipe, nil)
}

func (x *XrayCore) StartAsUser(stdinPipe []byte, username string) error {
	u, err := user.Lookup(username)
	if err != nil {
		return fmt.Errorf("failed to lookup user %s: %w", username, err)
	}

	uid, err := strconv.ParseUint(u.Uid, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid uid %s: %w", u.Uid, err)
	}

	gid, err := strconv.ParseUint(u.Gid, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid gid %s: %w", u.Gid, err)
	}

	cred := &syscall.Credential{
		Uid: uint32(uid),
		Gid: uint32(gid),
	}

	return x.startWithCredential(stdinPipe, cred)
}

func (x *XrayCore) startWithCredential(stdinPipe []byte, cred *syscall.Credential) error {
	x.mu.Lock()
	defer x.mu.Unlock()

	if x.running {
		return fmt.Errorf("command is already running")
	}

	xraybin, err := utils.GetXrayBin()
	if err != nil {
		return fmt.Errorf("failed to start xray %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, xraybin, "run")

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig:  syscall.SIGTERM,
		Credential: cred,
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("failed to get stdin %w", err)
	}

	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()

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
					x.channel_closed = true
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
			log.Println("error killing process", err)
		}
	}
	if x.cancel != nil {
		x.cancel()
	}
	x.running = false
}

func (x *XrayCore) IsRunning() bool {
	x.mu.Lock()
	defer x.mu.Unlock()
	return x.running
}
