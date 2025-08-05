package tunmode

import (
	// appconfig "bushuray-core/lib/AppConfig"
	"sync"
)

type TunModeManager struct {
	mu            sync.Mutex
	nekobox_core  NekoboxCore
	StatusChanged chan bool
	IsEnabled     bool
}

func (t *TunModeManager) Init() {
	t.nekobox_core = NekoboxCore{
		Exited: make(chan error),
	}
}

func (t *TunModeManager) Start() error {
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
