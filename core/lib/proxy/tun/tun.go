package tunmode

import (
	// appconfig "bushuray-core/lib/AppConfig"
	// "context"
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

func (t *TunModeManager) Start(proxy_ipv4 string, dns string) error {
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

	err = deleteIpRoutes(t.tun_name, t.tun_ip, interface_ip, proxy_ipv4, dns)
	if err != nil {
		log.Println("there was an error deleting ip routes", err)
	}

	err = setupIpRoutes(t.tun_name, t.tun_ip, interface_ip, proxy_ipv4, dns)
	if err != nil {
		log.Println("there was an error setting up ip routes", err)
	}

	log.Println("finished running ip commands")
	return nil
	// if t.nekobox_core.IsRunning() {
	// 	t.nekobox_core.Stop()
	// }
	//
	// t.nekobox_core = NekoboxCore{
	// 	Exited: make(chan error),
	// }
	//
	// if t.IsEnabled {
	// 	t.IsEnabled = false
	// 	t.StatusChanged <- t.IsEnabled
	// }
	//
	// if err := t.nekobox_core.Start(appconfig.GetConfig().SocksPort); err != nil {
	// 	return err
	// }
	//
	// t.IsEnabled = true
	// t.StatusChanged <- t.IsEnabled
	//
	// go func() {
	// 	for {
	// 		_, ok := <-t.nekobox_core.Exited
	// 		if !ok {
	// 			return
	// 		}
	// 		t.mu.Lock()
	// 		t.IsEnabled = false
	// 		t.StatusChanged <- t.IsEnabled
	// 		t.mu.Unlock()
	// 	}
	// }()
	//
	// return nil
}

func (t *TunModeManager) Stop() {
	// t.mu.Lock()
	// defer t.mu.Unlock()
	// t.nekobox_core.Stop()
	// t.IsEnabled = false
	// t.StatusChanged <- t.IsEnabled
}
