package cmd

import (
	"bushuray-core/lib/proxy/tun"
	"bushuray-core/structs"
	"log"
)

func (cmd *Cmd) DisableTun(data structs.DisableTunData, tun_manager *tunmode.TunModeManager) {
	tun_manager.Stop()
}

func (cmd *Cmd) EnableTun(data structs.EnableTunData, tun_manager *tunmode.TunModeManager) {
	err := tun_manager.Start()
	if err != nil {
		log.Println(err.Error())
		cmd.warn("enable-tun-failed", "Failed to enable tun mode")
		return
	}
}
