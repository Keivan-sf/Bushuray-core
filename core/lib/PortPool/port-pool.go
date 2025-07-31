package portpool

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

type PortPool struct {
	start_port int
	end_port   int
	in_use     map[int]bool
	mu         sync.Mutex
}

func CreatePortPool(start_port int, end_port int) *PortPool {
	if start_port < 0 || end_port < 0 || end_port <= start_port {
		log.Fatal("invalid testing port range")
	}
	return &PortPool{
		start_port: start_port,
		end_port:   end_port,
		in_use:     make(map[int]bool),
	}
}

func (p *PortPool) ReleasePort(port int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.in_use[port] = false
}

func (p *PortPool) GetPort() (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for i := p.start_port; i < p.end_port; i++ {
		if !p.isPortInUse(i) {
			p.in_use[i] = true
			return i, nil
		}
	}
	return -1, errors.New("no port availbe")
}

func (p *PortPool) isPortInUse(port int) bool {
	if p.in_use[port] {
		return true
	}

	tcp_addr := fmt.Sprintf(":%d", port)
	tcp_listener, err := net.Listen("tcp", tcp_addr)
	if err != nil {
		return true
	}
	tcp_listener.Close()

	udp_addr := fmt.Sprintf(":%d", port)
	udp_conn, err := net.ListenPacket("udp", udp_addr)
	if err != nil {
		return true
	}
	udp_conn.Close()

	return false
}
