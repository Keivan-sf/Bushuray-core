package mainproxy

import (
	"bushuray-core/lib"
	appconfig "bushuray-core/lib/AppConfig"
	portpool "bushuray-core/lib/PortPool"
	"bushuray-core/lib/proxy/xray"
	"bushuray-core/structs"
	"log"
	"sync"
)

// connect -> can also switch
// stop -> stops everything
// getStatus -> status
// test -> limit to 5 concurrent tests, but simple return interface

type ProxyManager struct {
	status            structs.ProxyStatus
	mu                sync.Mutex
	xray_core         xray.XrayCore
	StatusChanged     chan structs.ProxyStatus
	testChannel       chan structs.Profile
	TestResultChannel chan TestResult
	portPool          *portpool.PortPool
}

func (p *ProxyManager) Init() {
	p.status = structs.ProxyStatus{
		Connection: "disconnected",
	}
	p.StatusChanged = make(chan structs.ProxyStatus)
	test_channel := make(chan structs.Profile)
	go p.listenForTests(test_channel)
	p.testChannel = test_channel
	p.TestResultChannel = make(chan TestResult)
	p.xray_core = xray.XrayCore{
		Exited: make(chan error),
	}
	test_port_range := appconfig.GetConfig().TestPortRange
	p.portPool = portpool.CreatePortPool(test_port_range.Start, test_port_range.End)
}

func (p *ProxyManager) Connect(profile structs.Profile) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	app_config := appconfig.GetConfig()
	xray_config, err := lib.ParseUri(profile.Uri, app_config.SocksPort, app_config.HttpPort)
	if err != nil {
		return err
	}

	if p.xray_core.IsRunning() {
		p.xray_core.Stop()
	}

	p.xray_core = xray.XrayCore{
		Exited: make(chan error),
	}

	if p.status.Connection == "connected" {
		p.status = structs.ProxyStatus{
			Connection: "disconnected",
		}
		p.StatusChanged <- p.status
	}

	if err := p.xray_core.Start(xray_config); err != nil {
		return err
	}

	p.status = structs.ProxyStatus{
		Connection: "connected",
		Profile:    profile,
	}
	p.StatusChanged <- p.status
	log.Println("changing connection status to", p.status.Connection)

	go func() {
		for {
			_, ok := <-p.xray_core.Exited
			if !ok {
				return
			}
			p.mu.Lock()
			p.status = structs.ProxyStatus{
				Connection: "disconnected",
			}
			p.mu.Unlock()
			p.StatusChanged <- p.status
		}
	}()

	return nil
}

func (p *ProxyManager) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.xray_core.Stop()
	p.status = structs.ProxyStatus{
		Connection: "disconnected",
	}
	p.StatusChanged <- p.status
}

func (p *ProxyManager) GetStatus() structs.ProxyStatus {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.status
}
