package tunmode

import (
	"bushuray-core/utils"
	"context"
	"fmt"
	"log"
	"os/exec"
	"sync"
)

type Tun2Socks struct {
	mu             sync.Mutex
	cmd            *exec.Cmd
	cancel         context.CancelFunc
	running        bool
	channel_closed bool
	Exited         chan error
}

func (n *Tun2Socks) Start(tun_name string, port int) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.running {
		return fmt.Errorf("command is already running")
	}

	tun2socksbin, err := utils.GetTun2socksBin()
	if err != nil {
		return fmt.Errorf("failed to start tun: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	socks_proxy := fmt.Sprintf("socks5://127.0.0.1:%d", port)
	cmd := exec.CommandContext(ctx, tun2socksbin, "-device", tun_name, "-proxy", socks_proxy)

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
					n.channel_closed = true
				}
				n.running = false
			default:
				// Channel is full or no reader, don't block
			}
		}

	}()

	return nil
}

func (n *Tun2Socks) Stop() {
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

func (n *Tun2Socks) IsRunning() bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.running
}
