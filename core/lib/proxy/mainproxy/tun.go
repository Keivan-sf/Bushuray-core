package mainproxy

import tunscripts "bushuray-core/lib/proxy/mainproxy/scripts"

func (p *ProxyManager) prepareTunMode() error {
	return tunscripts.GenerateBxrayUser()
}

func (p *ProxyManager) enableTun() error {
	steps := []func() error{
		tunscripts.GenerateRoutingTable,
		tunscripts.BypassLocalAndLan,
		tunscripts.MarkAndForwardPackets,
		tunscripts.ApplyMainRules,
		tunscripts.ProxyGateway,
		tunscripts.LoosenRpFilters,
		tunscripts.BypassDns,
	}

	for _, step := range steps {
		if err := step(); err != nil {
			tunscripts.CleanUp()
			return err
		}
	}
	return nil
}

func (p *ProxyManager) disableTun() {
	tunscripts.CleanUp()
}
