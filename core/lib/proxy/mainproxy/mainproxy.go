package mainproxy

import (
	"fmt"
	"log"
	"sync"

	portpool "bushuray-core/lib/PortPool"
	"bushuray-core/lib/config"
	"bushuray-core/lib/proxy/builder"
	"bushuray-core/lib/proxy/xray"
	"bushuray-core/structs"
	"bushuray-core/utils"
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
		Exited: make(chan error, 1),
	}
	test_port_range := appConfig.TestPortRange
	p.portPool = portpool.CreatePortPool(test_port_range.Start, test_port_range.End)
	if err := p.recoverPersistedTunMode(); err != nil {
		log.Println("failed to recover persisted TUN state:", err)
	}
}

func (p *ProxyManager) ChangeTunMode(tun_mode bool) error {
	status := p.GetStatus()
	if tun_mode == status.IsTunEnabled {
		return nil
	}
	if status.Connection != "connected" {
		return fmt.Errorf("No profile is currently active")
	}

	if err := p.Stop(); err != nil {
		return err
	}
	return p.Connect(status.Profile, tun_mode)
}

func (p *ProxyManager) Connect(profile structs.Profile, tun_mode bool) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.status.IsTunEnabled || tunStateExists() {
		if err := p.disableTun(); err != nil {
			return fmt.Errorf("disable previous TUN mode: %w", err)
		}
	}
	if p.xray_core.IsRunning() {
		p.xray_core.Stop()
	}

	p.xray_core = xray.XrayCore{
		Exited: make(chan error, 1),
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

	xrayCoreConfig, err := utils.ParseUri(
		profile.Uri,
		p.appConfig.SocksPort,
		p.appConfig.HttpPort,
		tproxy_port,
	)
	if err != nil {
		return err
	}

	builder := builder.NewBuilder(xrayCoreConfig)
	if err := builder.ApplyDNS(p.appConfig.Dns); err != nil {
		return err
	}
	xrayConfig := builder.Build()

	if tun_mode {
		log.Println("enabling TUN with Xray TProxy on port 13345; no TUN network interface will be created")
		err = p.prepareTunMode()
		if err != nil {
			return err
		}
		if err := p.xray_core.StartAsUser(xrayConfig, "bxray_tproxy"); err != nil {
			return err
		}
		log.Println("TUN Xray process started; configuring policy routing")
		err = p.enableTun()
		if err != nil {
			p.xray_core.Stop()
			return fmt.Errorf("enable TUN mode: %w", err)
		}
		log.Println("successfully enabled tun")
	} else {
		if err := p.xray_core.Start(xrayConfig); err != nil {
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

	exited := p.xray_core.Exited
	go func(exited <-chan error) {
		for {
			exitErr, ok := <-exited
			if !ok {
				return
			}
			p.mu.Lock()
			if p.xray_core.Exited != exited {
				p.mu.Unlock()
				return
			}
			if p.status.IsTunEnabled || tunStateExists() {
				if err := p.disableTun(); err != nil {
					log.Println("failed to clean TUN mode after Xray exit:", err)
				}
			}
			p.status = structs.ProxyStatus{
				IsTunEnabled: false,
				Connection:   "disconnected",
			}
			status := p.status
			p.mu.Unlock()
			p.StatusChanged <- status
			if exitErr != nil {
				log.Println("Xray exited:", exitErr)
			}
		}
	}(exited)

	return nil
}

func (p *ProxyManager) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.xray_core.Stop()
	p.status = structs.ProxyStatus{
		IsTunEnabled: false,
		Connection:   "disconnected",
	}
	p.StatusChanged <- p.status

	var cleanupErr error
	if p.status.IsTunEnabled || tunStateExists() {
		cleanupErr = p.disableTun()
	}
	return wrapRollbackError(cleanupErr)
}

func (p *ProxyManager) GetStatus() structs.ProxyStatus {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.status
}
