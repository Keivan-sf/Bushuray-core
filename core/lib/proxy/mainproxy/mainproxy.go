package mainproxy

import (
	"bushuray-core/lib"
	portpool "bushuray-core/lib/PortPool"
	"bushuray-core/lib/config"
	"bushuray-core/lib/proxy/xray"
	"bushuray-core/structs"
	"fmt"
	"log"
	"sync"
)

type ProxyManager struct {
	status            structs.ProxyStatus
	mu                sync.Mutex
	appConfig         config.AppConfig
	xray_core         xray.XrayCore
	StatusChanged     chan structs.ProxyStatus
	testChannel       chan structs.Profile
	TestResultChannel chan TestResult
	portPool          *portpool.PortPool
	CurrentProfile    structs.Profile
}

func (p *ProxyManager) Init(appConfig config.AppConfig) {
	p.status = structs.ProxyStatus{
		Connection:   "disconnected",
		IsTunEnabled: false,
	}
	p.appConfig = appConfig
	p.StatusChanged = make(chan structs.ProxyStatus)
	test_channel := make(chan structs.Profile)
	go p.listenForTests(test_channel)
	p.testChannel = test_channel
	p.TestResultChannel = make(chan TestResult)
	p.xray_core = xray.XrayCore{
		Exited: make(chan error),
	}
	test_port_range := appConfig.TestPortRange
	p.portPool = portpool.CreatePortPool(test_port_range.Start, test_port_range.End)
}

func (p *ProxyManager) ChangeTunMode(tun_mode bool) error {
	if (tun_mode && p.status.IsTunEnabled) || (!tun_mode && !p.status.IsTunEnabled) {
		return nil
	}
	if p.GetStatus().Connection != "connected" {
		return fmt.Errorf("No profile is currently active")
	}
	current_profile := p.CurrentProfile

	p.Stop()
	return p.Connect(current_profile, tun_mode)
}

func (p *ProxyManager) Connect(profile structs.Profile, tun_mode bool) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.CurrentProfile = profile
	if p.xray_core.IsRunning() {
		p.xray_core.Stop()
	}

	p.xray_core = xray.XrayCore{
		Exited: make(chan error),
	}

	if p.status.Connection == "connected" {
		p.status = structs.ProxyStatus{
			Connection:   "disconnected",
			IsTunEnabled: false,
		}
		p.StatusChanged <- p.status
	}

	tproxy_port := -1
	if tun_mode {
		tproxy_port = 13345
	}

	xray_config, err := lib.ParseUri(
		profile.Uri,
		p.appConfig.SocksPort,
		p.appConfig.HttpPort,
		tproxy_port,
	)
	if err != nil {
		return err
	}

	if tun_mode {
		err = p.prepareTunMode()
		if err != nil {
			return err
		}
		if err := p.xray_core.StartAsUser(xray_config, "bxray_tproxy"); err != nil {
			return err
		}
		err = p.enableTun()
		if err != nil {
			log.Println("there was a problem enabling tun")
		} else {
			log.Println("successfully enabled tun")
		}
	} else {
		if err := p.xray_core.Start(xray_config); err != nil {
			return err
		}
	}

	p.status = structs.ProxyStatus{
		Connection:   "connected",
		IsTunEnabled: tun_mode,
		Profile:      profile,
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
				IsTunEnabled: false,
				Connection:   "disconnected",
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
	p.disableTun()
	p.status = structs.ProxyStatus{
		IsTunEnabled: false,
		Connection:   "disconnected",
	}
	p.StatusChanged <- p.status
}

func (p *ProxyManager) GetStatus() structs.ProxyStatus {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.status
}
