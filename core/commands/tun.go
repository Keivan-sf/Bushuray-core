package cmd

import (
	proxy "bushuray-core/lib/proxy/mainproxy"
	"bushuray-core/lib/proxy/tun"
	"bushuray-core/structs"
	"bushuray-core/utils"
	"errors"
	"fmt"
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

	resolved, err := resolveHostAndAddress(status.Profile)
	if err != nil {
		log.Println(err)
		cmd.warn("enable-tun-failed", "failed to resolve profile host")
		return
	}

	log.Println("resolved:", resolved)

	err = tun_manager.Start(resolved, "8.8.8.8")
	if err != nil {
		log.Println(err.Error())
		cmd.warn("enable-tun-failed", "Failed to enable tun mode")
		return
	}
}

func resolveHostAndAddress(profile structs.Profile) ([]string, error) {
	var ipv4s []string
	var errs []error
	if profile.Host != "" {
		resolved, err := utils.ResolveDomainIpv4(profile.Host)
		if err == nil {
			ipv4s = append(ipv4s, resolved...)
		} else {
			errs = append(errs, err)
		}
	}
	if profile.Address != "" {
		resolved, err := utils.ResolveDomainIpv4(profile.Address)
		if err == nil {
			ipv4s = append(ipv4s, resolved...)
		} else {
			errs = append(errs, err)
		}
	}
	if len(ipv4s) == 0 {
		return ipv4s, fmt.Errorf("failed to resolve any ipv4s: %w", errors.Join(errs...))
	}
	log.Println("at the end2:", ipv4s)
	return ipv4s, nil
}
