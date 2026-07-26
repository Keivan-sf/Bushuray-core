package cmd

import (
	proxy "bushuray-core/lib/proxy/mainproxy"
	"bushuray-core/structs"
	"log"
)

func (cmd *Cmd) Disconnect(data structs.DisconnectData, proxy_manager *proxy.ProxyManager) {
	ConnectionMutex.Lock()
	defer ConnectionMutex.Unlock()

	proxy_manager.Stop()
}

func (cmd *Cmd) Connect(data structs.ConnectData, proxy_manager *proxy.ProxyManager) {
	ConnectionMutex.Lock()
	defer ConnectionMutex.Unlock()

	profile, err := cmd.DB.GetProfile(data.Profile.GroupId, data.Profile.Id)
	if err != nil {
		log.Println(err.Error())
		cmd.warn("connect-failed", "Failed to connect")
		return
	}
	was_tun_enabled := proxy_manager.GetStatus().IsTunEnabled

	if err := proxy_manager.Connect(profile, was_tun_enabled); err != nil {
		log.Println(err.Error())
		cmd.warn("connect-failed", "Failed to connect")
		return
	}

	err = cmd.DB.UpdateLatestConnectedProfile(profile.GroupId, profile.Id)
	if err != nil {
		log.Println(err.Error())
	}
}

func (cmd *Cmd) EnableTun(data structs.EnableTunData, proxy_manager *proxy.ProxyManager) {
	ConnectionMutex.Lock()
	defer ConnectionMutex.Unlock()

	if err := proxy_manager.ChangeTunMode(true); err != nil {
		log.Println(err.Error())
		cmd.warn("enable-tun-failed", "Failed to enable tun mode")
	}
}

func (cmd *Cmd) DisableTun(data structs.DisableTunData, proxy_manager *proxy.ProxyManager) {
	ConnectionMutex.Lock()
	defer ConnectionMutex.Unlock()

	if err := proxy_manager.ChangeTunMode(false); err != nil {
		log.Println(err.Error())
		cmd.warn("disable-tun-failed", "Failed to reconnect")
	}
}
