package tunmode

import (
	// appconfig "bushuray-core/lib/AppConfig"
	// "context"
	// appconfig "bushuray-core/lib/AppConfig"
	appconfig "bushuray-core/lib/AppConfig"
	"errors"
	"fmt"
	"log"
	"sync"
)

type TunModeManager struct {
	mu                   sync.Mutex
	tun2socks         Tun2Socks
	tun_name             string
	tun_ip               string
	default_interface    string
	default_interface_ip string
	proxy_ipv4s          []string
	dns                  string
	StatusChanged        chan bool
	IsEnabled            bool
}

func (t *TunModeManager) Init() {
	t.tun_name = "bushuraytun"
	t.tun_ip = "198.18.0.1"
	t.StatusChanged = make(chan bool)
	t.tun2socks = Tun2Socks{
		Exited: make(chan error),
	}
}

func (t *TunModeManager) Start(proxy_ipv4s []string, dns string) error {
	log.Println("running ip commands")
	interface_name, interface_ip, err := GetDefaultInterfaceAndIP()
	if err != nil {
		return fmt.Errorf("failed on getting default interafce %w", err)
	}

	t.default_interface_ip = interface_ip
	t.default_interface = interface_name
	t.proxy_ipv4s = proxy_ipv4s
	t.dns = dns

	err = t.clearNetworkRules()
	if err != nil {
		return fmt.Errorf("there was an error clearing network rules %w", err)
	}

	err = setupDnsHijackRules(t.default_interface, t.default_interface_ip, t.dns)
	if err != nil {
		return fmt.Errorf("there was an error setting up dns hijack rules %w", err)
	}

	err = createTun(t.tun_name, t.tun_ip)
	if err != nil {
		t.clearNetworkRules()
		return fmt.Errorf("there was an error creating tun interface %w", err)
	}

	err = setupProxyIpRoutes(t.proxy_ipv4s, t.default_interface_ip)
	if err != nil {
		t.clearNetworkRules()
		return fmt.Errorf("there was an error setting up proxy ip routes %w", err)
	}

	err = setupDnsIpRoute(t.dns, t.default_interface_ip)
	if err != nil {
		t.clearNetworkRules()
		return fmt.Errorf("there was an error setting up dns ip route %w", err)
	}

	err = setupTunIpRoute(t.tun_name, t.tun_ip)
	if err != nil {
		t.clearNetworkRules()
		return fmt.Errorf("there was an error setting up tun ip route %w", err)
	}

	log.Println("finished running ip commands")
	if t.tun2socks.IsRunning() {
		t.tun2socks.Stop()
	}

	t.tun2socks = Tun2Socks{
		Exited: make(chan error),
	}

	if t.IsEnabled {
		t.IsEnabled = false
		t.StatusChanged <- t.IsEnabled
	}

	if err := t.tun2socks.Start(t.tun_name, appconfig.GetConfig().SocksPort); err != nil {
		return err
	}

	t.IsEnabled = true
	t.StatusChanged <- t.IsEnabled

	go func() {
		for {
			_, ok := <-t.tun2socks.Exited
			if !ok {
				return
			}
			t.mu.Lock()
			t.IsEnabled = false
			t.StatusChanged <- t.IsEnabled
			t.mu.Unlock()
		}
	}()

	return nil
}

func (t *TunModeManager) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.clearNetworkRules()
	t.tun2socks.Stop()
	t.IsEnabled = false
	t.StatusChanged <- t.IsEnabled
}

func (t *TunModeManager) clearNetworkRules() error {
	errs := []error{
		deleteTunIpRoute(t.tun_name, t.tun_ip),
		deleteTun(t.tun_name),
		deleteDnsIpRoute(t.dns, t.default_interface_ip),
		cleanDnsHijackRules(t.default_interface, t.default_interface_ip, t.dns),
		deleteProxyIpRoutes(t.proxy_ipv4s, t.default_interface_ip),
	}
	return errors.Join(errs...)
}
