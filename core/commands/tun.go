package cmd

import (
	proxy "bushuray-core/lib/proxy/mainproxy"
	"bushuray-core/lib/proxy/tun"
	"bushuray-core/structs"
	"bushuray-core/utils"
	"log"
)

func (cmd *Cmd) DisableTun(data structs.DisableTunData, tun_manager *tunmode.TunModeManager) {
	tun_manager.Stop()
}

func (cmd *Cmd) EnableTun(data structs.EnableTunData, proxy_manager *proxy.ProxyManager, tun_manager *tunmode.TunModeManager) {
	log.Println("on enable tun")
	status := proxy_manager.GetStatus()
	if status.Connection != "connected" {
		cmd.warn("enable-tun-failed", "A profile must be connected for tun mode to operate")
		return
	}
	host := status.Profile.Host
	address := status.Profile.Host
	log.Println("host and address are:", host, address)
	endpoint := ""
	if host != "" {
		endpoint = status.Profile.Host
	} else {
		endpoint = status.Profile.Address
	}
	if endpoint == "" {
		cmd.warn("enable-tun-failed", "no endpoint found for the connected profile")
		return
	}
	log.Println("the endpoint is:", endpoint)
	resolved, err := utils.ResolveDomain(endpoint)
	if err != nil {
		cmd.warn("enable-tun-failed", "failed to resolve connected end-point")
		return
	}
	log.Println("resolved:", resolved)

	// err := tun_manager.Start()
	// if err != nil {
	// 	log.Println(err.Error())
	// 	cmd.warn("enable-tun-failed", "Failed to enable tun mode")
	// 	return
	// }
}
