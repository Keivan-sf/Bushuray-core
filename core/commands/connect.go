package cmd

import (
	"bushuray-core/lib/proxy"
	"bushuray-core/structs"
	"log"
)

func (cmd *Cmd) Connect(data structs.ConnectData, proxy_manager *proxy.ProxyManager) {
	profile, err := cmd.DB.GetProfile(data.Profile.GroupId, data.Profile.Id)
	if err != nil {
		log.Println(err.Error())
		return
		// warn client
	}

	if err := proxy_manager.Connect(profile); err != nil {
		log.Println(err.Error())
		return
		// warn client
	}
}
