package cmd

import (
	proxy "bushuray-core/lib/proxy/mainproxy"
	tunmode "bushuray-core/lib/proxy/tun"
	"bushuray-core/structs"
	"log"
)

func (cmd *Cmd) Disconnect(data structs.DisconnectData, proxy_manager *proxy.ProxyManager, tun_manager *tunmode.TunModeManager) {
	proxy_manager.Stop()
	tun_manager.Stop()
}

func (cmd *Cmd) Connect(data structs.ConnectData, proxy_manager *proxy.ProxyManager, tun_manager *tunmode.TunModeManager) {
	profile, err := cmd.DB.GetProfile(data.Profile.GroupId, data.Profile.Id)
	if err != nil {
		log.Println(err.Error())
		cmd.warn("connect-failed", "Failed to connect")
		return
	}
	was_tun_enabled := tun_manager.IsEnabled

	if was_tun_enabled {
		tun_manager.Stop()
	}

	if err := proxy_manager.Connect(profile); err != nil {
		log.Println(err.Error())
		cmd.warn("connect-failed", "Failed to connect")
		return
	}

	if was_tun_enabled {
		cmd.enableTun(profile, tun_manager)
	}
}
