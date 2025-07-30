package portpool

import (
	"errors"
	"fmt"
	"net"
)

type PortPool struct {
	start_port int
	end_port   int
	in_use     map[int]bool
}

func (p *PortPool) ReleasePort(port int) {
	p.in_use[port] = false
}

func (p *PortPool) GetPort() (int, error) {
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
