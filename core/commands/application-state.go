package cmd

import (
	proxy "bushuray-core/lib/proxy/mainproxy"
	"bushuray-core/structs"
	"log"
)

func (cmd *Cmd) GetApplicationState(data structs.GetApplicationStateData, proxy_manager *proxy.ProxyManager) {
	groups, err := cmd.DB.GetAllGroupsAndProfiles()
	if err != nil {
		log.Println(err.Error())
		return
	}

	application_state := structs.ApplicationState{
		Groups:           groups,
		ConnectionStatus: proxy_manager.GetStatus(),
	}

	cmd.send("application-state", application_state)
}
