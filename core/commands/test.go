package cmd

import (
	proxy "bushuray-core/lib/proxy/mainproxy"
	"bushuray-core/structs"
	"log"
)

func (cmd *Cmd) TestProfile(data structs.TestProfileData, proxy_manager *proxy.ProxyManager) {
	profile, err := cmd.DB.GetProfile(data.Profile.GroupId, data.Profile.Id)
	if err != nil {
		log.Println(err.Error())
		return
	}
	proxy_manager.TestProfile(profile)
}
