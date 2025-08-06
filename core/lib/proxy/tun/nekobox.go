package tunmode

import (
	"bushuray-core/lib"
	"context"
	"fmt"
	"log"
	"os/exec"
	"path"
	"sync"
)

type NekoboxCore struct {
	mu             sync.Mutex
	cmd            *exec.Cmd
	cancel         context.CancelFunc
	running        bool
	channel_closed bool
	Exited         chan error
}

func (n *NekoboxCore) Start(tun_name string, port int) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.running {
		return fmt.Errorf("command is already running")
	}

	ctx, cancel := context.WithCancel(context.Background())

	nekoboxbin := path.Join(lib.GetWorkingDir(), "bin", "tun2socks")
	socks_proxy := fmt.Sprintf("socks5://127.0.0.1:%d", port)
	// ./tun2socks-linux-amd64 -device bushuraytun -proxy socks5://127.0.0.1:3090
	cmd := exec.CommandContext(ctx, nekoboxbin, "-device", tun_name, "-proxy", socks_proxy)

	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		cancel()
		return err
	}

	n.cmd = cmd
	n.cancel = cancel
	n.running = true

	go func() {
		err := cmd.Wait()
		n.mu.Lock()
		defer n.mu.Unlock()
		if ctx.Err() == nil {
			select {
			case n.Exited <- err:
				if n.running {
					close(n.Exited)
				}
				n.running = false
			default:
				// Channel is full or no reader, don't block
			}
		}

	}()

	return nil
}

func (n *NekoboxCore) Stop() {
	n.mu.Lock()
	defer n.mu.Unlock()
	if !n.channel_closed {
		close(n.Exited)
		n.channel_closed = true
	}
	if n.cmd != nil && n.cmd.Process != nil {
		err := n.cmd.Process.Kill()
		if err != nil {
			log.Println("error killing proces", err)
		}
	}
	if n.cancel != nil {
		n.cancel()
	}
	n.running = false
}

func (n *NekoboxCore) IsRunning() bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.running
}
