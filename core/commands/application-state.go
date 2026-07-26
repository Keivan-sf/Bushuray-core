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
		cmd.warn("read-application-state-failed", "failed to read application state")
		return
	}

	proxyStatus := proxy_manager.GetStatus()
	application_state := structs.ApplicationState{
		Groups:           groups,
		ConnectionStatus: proxyStatus,
		TunStatus:        proxyStatus.IsTunEnabled,
	}

	cmd.send("application-state", application_state)
}
