package tunmode

import (
	// appconfig "bushuray-core/lib/AppConfig"
	// "context"
	"log"
	"net"
	"sync"
)

type TunModeManager struct {
	mu            sync.Mutex
	nekobox_core  NekoboxCore
	tun_name      string
	StatusChanged chan bool
	IsEnabled     bool
}

func (t *TunModeManager) Init() {
	t.tun_name = "bushuraytun"
	t.nekobox_core = NekoboxCore{
		Exited: make(chan error),
	}
}

func (t *TunModeManager) Start(ips []net.IP, dns string) error {
	// ctx, cancel := context.WithCancel(context.Background())
	log.Println("running ip commands")
	iname, ip, err := GetDefaultInterfaceAndIP()
	log.Println("default interface be like:", iname, ip)
	if err != nil {
		log.Println("failed on getting default interafce", err)
		return nil
	}
	err = cleanDnsHijackRules(iname, ip, dns)
	if err != nil {
		log.Println("there was an error cleaning dns hijack rules", err)
	}
	err = setupDnsHijackRules(iname, ip, dns)
	if err != nil {
		log.Println("there was an error setting up dns hijack rules", err)
	}
	log.Println("finished running ip table commands")

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
