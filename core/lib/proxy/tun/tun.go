package tunmode

import (
	// appconfig "bushuray-core/lib/AppConfig"
	// "context"
	// appconfig "bushuray-core/lib/AppConfig"
	appconfig "bushuray-core/lib/AppConfig"
	"log"
	"sync"
)

type TunModeManager struct {
	mu            sync.Mutex
	nekobox_core  NekoboxCore
	tun_name      string
	tun_ip        string
	StatusChanged chan bool
	IsEnabled     bool
}

func (t *TunModeManager) Init() {
	t.tun_name = "bushuraytun"
	t.tun_ip = "198.18.0.1"
	t.nekobox_core = NekoboxCore{
		Exited: make(chan error),
	}
}

func (t *TunModeManager) Start(proxy_ipv4s []string, dns string) error {
	// ctx, cancel := context.WithCancel(context.Background())
	log.Println("running ip commands")
	interface_name, interface_ip, err := GetDefaultInterfaceAndIP()
	if err != nil {
		log.Println("failed on getting default interafce", err)
		return nil
	}

	err = cleanDnsHijackRules(interface_name, interface_ip, dns)
	if err != nil {
		log.Println("there was an error cleaning dns hijack rules", err)
	}

	err = setupDnsHijackRules(interface_name, interface_ip, dns)
	if err != nil {
		log.Println("there was an error setting up dns hijack rules", err)
	}

	err = deleteTun(t.tun_name)
	if err != nil {
		log.Println("there was an error deleting tun interface", err)
	}

	err = createTun(t.tun_name, t.tun_ip)
	if err != nil {
		log.Println("there was an error creating tun interface", err)
	}

	err = deleteProxyIpRoutes(proxy_ipv4s, interface_ip)
	if err != nil {
		log.Println("there was an error deleting proxy ip routes", err)
	}

	err = setupProxyIpRoutes(proxy_ipv4s, interface_ip)
	if err != nil {
		log.Println("there was an error setting up proxy ip routes", err)
	}

	err = deleteDnsIpRoute(dns, interface_ip)
	if err != nil {
		log.Println("there was an error deleting dns ip route", err)
	}

	err = setupDnsIpRoute(dns, interface_ip)
	if err != nil {
		log.Println("there was an error setting up dns ip route", err)
	}

	err = deleteTunIpRoute(t.tun_name, t.tun_ip)
	if err != nil {
		log.Println("there was an error deleting tun ip route", err)
	}

	err = setupTunIpRoute(t.tun_name, t.tun_ip)
	if err != nil {
		log.Println("there was an error setting up tun ip route", err)
	}

	log.Println("finished running ip commands")
	if t.nekobox_core.IsRunning() {
		t.nekobox_core.Stop()
	}

	t.nekobox_core = NekoboxCore{
		Exited: make(chan error),
	}

	if t.IsEnabled {
		t.IsEnabled = false
		t.StatusChanged <- t.IsEnabled
	}

	if err := t.nekobox_core.Start(t.tun_name, appconfig.GetConfig().SocksPort); err != nil {
		return err
	}

	t.IsEnabled = true
	t.StatusChanged <- t.IsEnabled

	go func() {
		for {
			_, ok := <-t.nekobox_core.Exited
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
	// t.mu.Lock()
	// defer t.mu.Unlock()
	// t.nekobox_core.Stop()
	// t.IsEnabled = false
	// t.StatusChanged <- t.IsEnabled
}
