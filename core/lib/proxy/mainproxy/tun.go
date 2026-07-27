package mainproxy

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"bushuray-core/lib/config"
	tunscripts "bushuray-core/lib/proxy/mainproxy/scripts"
)

func (p *ProxyManager) prepareTunMode() error {
	return tunscripts.GenerateBxrayUser()
}

func (p *ProxyManager) enableTun() error {
	log.Println("TUN setup: wait for Xray TCP/UDP listener on port 13345")
	listener, err := tunscripts.WaitForTProxyListener(13345, 3*time.Second)
	if err != nil {
		return err
	}
	log.Println("TUN listener ready:", strings.TrimSpace(listener))

	if err := p.beginTunMode(); err != nil {
		return err
	}

	type tunStep struct {
		name string
		run  func() error
	}
	steps := []tunStep{
		{name: "generate routing table", run: tunscripts.GenerateRoutingTable},
		{name: "create TProxy chain", run: tunscripts.BypassLocalAndLan},
		{name: "configure TProxy targets", run: tunscripts.MarkAndForwardPackets},
		{name: "attach PREROUTING chain", run: tunscripts.ApplyMainRules},
		{name: "configure OUTPUT routing", run: tunscripts.ProxyGateway},
		{name: "loosen rp_filter", run: tunscripts.LoosenRpFilters},
	}
	if p.appConfig.Dns.Mode == config.DNSModeSystem {
		steps = append(steps, tunStep{name: "bypass system DNS", run: tunscripts.BypassDns})
	} else {
		steps = append(steps, tunStep{name: "hijack DNS", run: tunscripts.HijackDns})
	}

	for _, step := range steps {
		log.Println("TUN setup:", step.name)
		if err := step.run(); err != nil {
			rollbackErr := p.rollbackTunMode()
			return errors.Join(
				fmt.Errorf("%s: %w", step.name, err),
				wrapRollbackError(rollbackErr),
			)
		}
	}

	log.Println("TUN setup: verify policy routing, netfilter hooks and Xray listener")
	diagnostics, err := tunscripts.WaitForTunReady(13345, 3*time.Second)
	if err != nil {
		rollbackErr := p.rollbackTunMode()
		return errors.Join(err, wrapRollbackError(rollbackErr))
	}
	log.Println("TUN ready:", strings.TrimSpace(diagnostics))
	return nil
}

func (p *ProxyManager) disableTun() error {
	return p.rollbackTunMode()
}

func wrapRollbackError(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("roll back TUN mode: %w", err)
}
