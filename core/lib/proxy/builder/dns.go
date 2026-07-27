package builder

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"

	"bushuray-core/lib/config"
)

const (
	xrayDNSTag            = "bushuray-dns"
	xrayDNSRuleTag        = "bushuray-dns-routing"
	xrayDNSHijackRuleTag  = "bushuray-dns-hijack-routing"
	xrayDNSOutboundTag    = "bushuray-dns-out"
	xrayDirectOutboundTag = "bushuray-dns-direct"
	xrayProxyOutboundTag  = "proxy"
)

func (b *Builder) ApplyDNS(dnsConfig config.DNSConfig) error {
	if err := validateDNSConfig(dnsConfig); err != nil {
		return err
	}

	var coreConfig map[string]any
	if err := json.Unmarshal(b.coreJSON, &coreConfig); err != nil {
		return fmt.Errorf("parse Xray config: %w", err)
	}
	if coreConfig == nil {
		return ErrCoreIsNill
	}

	servers := dnsConfig.Servers
	if dnsConfig.Mode == config.DNSModeSystem {
		servers = []string{"localhost"}
	}

	dns := map[string]any{
		"servers":       servers,
		"queryStrategy": dnsConfig.QueryStrategy,
	}

	switch dnsConfig.Mode {
	case config.DNSModeSystem:
		removeDNSRoutingRules(coreConfig)
		removeDNSOutbound(coreConfig)
	case config.DNSModeProxy:
		dns["tag"] = xrayDNSTag
		proxyTag, err := ensureProxyOutboundTag(coreConfig)
		if err != nil {
			return err
		}
		dnsOutboundTag, err := ensureDNSOutbound(coreConfig, dnsConfig, proxyTag)
		if err != nil {
			return err
		}
		applyDNSRoutingRules(coreConfig, proxyTag, dnsOutboundTag)
	case config.DNSModeDirect:
		dns["tag"] = xrayDNSTag
		directTag, err := ensureDirectOutbound(coreConfig)
		if err != nil {
			return err
		}
		dnsOutboundTag, err := ensureDNSOutbound(coreConfig, dnsConfig, "")
		if err != nil {
			return err
		}
		applyDNSRoutingRules(coreConfig, directTag, dnsOutboundTag)
	}

	coreConfig["dns"] = dns

	coreJSON, err := json.Marshal(coreConfig)
	if err != nil {
		return fmt.Errorf("marshal Xray config with DNS: %w", err)
	}
	b.coreJSON = coreJSON
	return nil
}

func validateDNSConfig(dnsConfig config.DNSConfig) error {
	switch dnsConfig.Mode {
	case config.DNSModeSystem, config.DNSModeProxy, config.DNSModeDirect:
	default:
		return fmt.Errorf("unsupported DNS mode %q", dnsConfig.Mode)
	}

	switch dnsConfig.QueryStrategy {
	case config.DNSUseIP, config.DNSUseIPv4, config.DNSUseIPv6, config.DNSUseSystem:
	default:
		return fmt.Errorf("unsupported DNS query strategy %q", dnsConfig.QueryStrategy)
	}

	if dnsConfig.Mode == config.DNSModeSystem {
		return nil
	}
	if len(dnsConfig.Servers) == 0 {
		return errors.New("DNS servers cannot be empty in proxy or tun mode")
	}
	for _, server := range dnsConfig.Servers {
		server = strings.TrimSpace(server)
		if server == "" {
			return errors.New("DNS server address cannot be empty")
		}
	}
	return nil
}

func ensureDNSOutbound(
	coreConfig map[string]any,
	dnsConfig config.DNSConfig,
	proxyTag string,
) (string, error) {
	outbounds, err := getOutbounds(coreConfig)
	if err != nil {
		return "", err
	}

	for _, outboundValue := range outbounds {
		outbound, ok := outboundValue.(map[string]any)
		if !ok || stringValue(outbound["tag"]) != xrayDNSOutboundTag {
			continue
		}
		if stringValue(outbound["protocol"]) != "dns" {
			return "", fmt.Errorf("outbound tag %q is already used by protocol %q",
				xrayDNSOutboundTag, stringValue(outbound["protocol"]))
		}
		configureDNSOutbound(outbound, dnsConfig, proxyTag)
		return xrayDNSOutboundTag, nil
	}

	outbound := map[string]any{
		"protocol": "dns",
		"tag":      xrayDNSOutboundTag,
	}
	configureDNSOutbound(outbound, dnsConfig, proxyTag)
	coreConfig["outbounds"] = append(outbounds, outbound)
	return xrayDNSOutboundTag, nil
}

func configureDNSOutbound(
	outbound map[string]any,
	dnsConfig config.DNSConfig,
	proxyTag string,
) {
	rules := []any{
		map[string]any{
			"action": "hijack",
			"qType":  "1,28",
		},
	}
	settings := map[string]any{"rules": rules}

	// The built-in DNS resolver handles A and AAAA. Forward other traditional
	// DNS record types to the first plain-IP upstream instead of returning an
	// empty response.
	if len(dnsConfig.Servers) > 0 {
		server := strings.TrimSpace(dnsConfig.Servers[0])
		if ip := net.ParseIP(server); ip != nil {
			settings["rewriteAddress"] = ip.String()
			settings["rules"] = append(rules, map[string]any{"action": "direct"})
		}
	}
	outbound["settings"] = settings

	if proxyTag != "" {
		outbound["proxySettings"] = map[string]any{"tag": proxyTag}
	} else {
		delete(outbound, "proxySettings")
	}
}

func removeDNSOutbound(coreConfig map[string]any) {
	outbounds, err := getOutbounds(coreConfig)
	if err != nil {
		return
	}
	filtered := make([]any, 0, len(outbounds))
	for _, outboundValue := range outbounds {
		outbound, ok := outboundValue.(map[string]any)
		if ok && stringValue(outbound["tag"]) == xrayDNSOutboundTag &&
			stringValue(outbound["protocol"]) == "dns" {
			continue
		}
		filtered = append(filtered, outboundValue)
	}
	coreConfig["outbounds"] = filtered
}

func ensureProxyOutboundTag(coreConfig map[string]any) (string, error) {
	outbounds, err := getOutbounds(coreConfig)
	if err != nil {
		return "", err
	}

	for _, outboundValue := range outbounds {
		outbound, ok := outboundValue.(map[string]any)
		if ok && stringValue(outbound["tag"]) == xrayProxyOutboundTag {
			return xrayProxyOutboundTag, nil
		}
	}

	if len(outbounds) == 0 {
		return "", errors.New("xray config has no proxy outbound")
	}
	firstOutbound, ok := outbounds[0].(map[string]any)
	if !ok {
		return "", errors.New("first Xray outbound is not an object")
	}
	if tag := stringValue(firstOutbound["tag"]); tag != "" {
		return tag, nil
	}

	tag := uniqueOutboundTag(outbounds, "bushuray-proxy")
	firstOutbound["tag"] = tag
	return tag, nil
}

func ensureDirectOutbound(coreConfig map[string]any) (string, error) {
	outbounds, err := getOutbounds(coreConfig)
	if err != nil {
		return "", err
	}

	for _, outboundValue := range outbounds {
		outbound, ok := outboundValue.(map[string]any)
		if !ok || stringValue(outbound["protocol"]) != "freedom" {
			continue
		}
		if tag := stringValue(outbound["tag"]); tag != "" {
			return tag, nil
		}
		tag := uniqueOutboundTag(outbounds, xrayDirectOutboundTag)
		outbound["tag"] = tag
		return tag, nil
	}

	tag := uniqueOutboundTag(outbounds, xrayDirectOutboundTag)
	coreConfig["outbounds"] = append(outbounds, map[string]any{
		"protocol": "freedom",
		"tag":      tag,
	})
	return tag, nil
}

func getOutbounds(coreConfig map[string]any) ([]any, error) {
	value, ok := coreConfig["outbounds"]
	if !ok {
		return nil, errors.New("xray config has no outbounds")
	}
	outbounds, ok := value.([]any)
	if !ok {
		return nil, errors.New("xray outbounds is not an array")
	}
	return outbounds, nil
}

func applyDNSRoutingRules(coreConfig map[string]any, outboundTag string, dnsOutboundTag string) {
	routing := getOrCreateObject(coreConfig, "routing")
	rules := getOrCreateArray(routing, "rules")

	filteredRules := make([]any, 0, len(rules)+2)
	filteredRules = append(filteredRules, map[string]any{
		"type":        "field",
		"ruleTag":     xrayDNSHijackRuleTag,
		"port":        "53",
		"network":     "tcp,udp",
		"outboundTag": dnsOutboundTag,
	})
	filteredRules = append(filteredRules, map[string]any{
		"type":        "field",
		"ruleTag":     xrayDNSRuleTag,
		"inboundTag":  []string{xrayDNSTag},
		"outboundTag": outboundTag,
	})
	for _, ruleValue := range rules {
		rule, ok := ruleValue.(map[string]any)
		if ok {
			tag := stringValue(rule["ruleTag"])
			if tag == xrayDNSRuleTag || tag == xrayDNSHijackRuleTag {
				continue
			}
		}
		filteredRules = append(filteredRules, ruleValue)
	}
	routing["rules"] = filteredRules
}

func removeDNSRoutingRules(coreConfig map[string]any) {
	routing, ok := coreConfig["routing"].(map[string]any)
	if !ok {
		return
	}
	rules, ok := routing["rules"].([]any)
	if !ok {
		return
	}

	filteredRules := make([]any, 0, len(rules))
	for _, ruleValue := range rules {
		rule, ok := ruleValue.(map[string]any)
		if ok {
			tag := stringValue(rule["ruleTag"])
			if tag == xrayDNSRuleTag || tag == xrayDNSHijackRuleTag {
				continue
			}
		}
		filteredRules = append(filteredRules, ruleValue)
	}
	routing["rules"] = filteredRules
}

func getOrCreateObject(parent map[string]any, key string) map[string]any {
	if object, ok := parent[key].(map[string]any); ok {
		return object
	}
	object := make(map[string]any)
	parent[key] = object
	return object
}

func getOrCreateArray(parent map[string]any, key string) []any {
	if array, ok := parent[key].([]any); ok {
		return array
	}
	array := make([]any, 0)
	parent[key] = array
	return array
}

func uniqueOutboundTag(outbounds []any, base string) string {
	usedTags := make(map[string]struct{}, len(outbounds))
	for _, outboundValue := range outbounds {
		outbound, ok := outboundValue.(map[string]any)
		if !ok {
			continue
		}
		if tag := stringValue(outbound["tag"]); tag != "" {
			usedTags[tag] = struct{}{}
		}
	}

	if _, exists := usedTags[base]; !exists {
		return base
	}
	for suffix := 2; ; suffix++ {
		candidate := fmt.Sprintf("%s-%d", base, suffix)
		if _, exists := usedTags[candidate]; !exists {
			return candidate
		}
	}
}

func stringValue(value any) string {
	text, _ := value.(string)
	return text
}
