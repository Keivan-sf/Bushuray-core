package cmd

import (
	proxy "bushuray-core/lib/proxy/mainproxy"
	"bushuray-core/structs"
	"log"
)

func (cmd *Cmd) Disconnect(data structs.DisconnectData, proxy_manager *proxy.ProxyManager) {
	proxy_manager.Stop()
}

func (cmd *Cmd) Connect(data structs.ConnectData, proxy_manager *proxy.ProxyManager) {
	profile, err := cmd.DB.GetProfile(data.Profile.GroupId, data.Profile.Id)
	if err != nil {
		log.Println(err.Error())
		cmd.warn("connect-failed", "Failed to connect")
		return
		// warn client
	}

	if err := proxy_manager.Connect(profile); err != nil {
		log.Println(err.Error())
		return
		// warn client
	}
}
