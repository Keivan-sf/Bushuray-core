package tunmode

import (
	"bushuray-core/lib"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
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

func (n *NekoboxCore) Start(port int) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.running {
		return fmt.Errorf("command is already running")
	}

	ctx, cancel := context.WithCancel(context.Background())

	tun_config := strings.Replace(json_config_template, "%SOCKSPORT%", strconv.Itoa(port), 1)
	tun_config_path := path.Join("/", "tmp", "bushuray-tun-config.json")

	os.WriteFile(tun_config_path, []byte(tun_config), 0777)

	nekoboxbin := path.Join(lib.GetWorkingDir(), "bin", "nekobox_core")

	cmd := exec.CommandContext(ctx, nekoboxbin, "run", "-c", tun_config_path)

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
