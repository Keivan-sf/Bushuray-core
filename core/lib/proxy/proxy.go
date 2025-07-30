package proxy

import (
	"bushuray-core/lib"
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
	status        structs.ProxyStatus
	mu            sync.Mutex
	xray_core     xray.XrayCore
	StatusChanged chan structs.ProxyStatus
}

func (p *ProxyManager) Init() {
	p.StatusChanged = make(chan structs.ProxyStatus)
	p.xray_core = xray.XrayCore{
		Exited: make(chan error),
	}
}

func (p *ProxyManager) Connect(profile structs.Profile) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	xray_config, err := lib.ParseUri(profile.Uri, 3090, 3091)
	if err != nil {
		return err
	}

	if p.xray_core.IsRunning() {
		p.xray_core.Stop()
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
		<-p.xray_core.Exited
		p.mu.Lock()
		p.status = structs.ProxyStatus{
			Connection: "disconnected",
		}
		p.mu.Unlock()
		p.StatusChanged <- p.status
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
