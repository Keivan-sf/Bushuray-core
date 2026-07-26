package builder

import (
	"encoding/json"
	"fmt"
)

// ApplyTProxyNetworkCompatibility keeps the tunnel inbound compatible with
// Xray releases from both sides of the network -> allowedNetwork rename.
// Unknown JSON fields are ignored by Xray, so providing both names is safe.
func (b *Builder) ApplyTProxyNetworkCompatibility() (bool, error) {
	var coreConfig map[string]any
	if err := json.Unmarshal(b.coreJSON, &coreConfig); err != nil {
		return false, fmt.Errorf("parse Xray config for TProxy compatibility: %w", err)
	}
	if coreConfig == nil {
		return false, ErrCoreIsNill
	}

	inbounds, ok := coreConfig["inbounds"].([]any)
	if !ok {
		return false, fmt.Errorf("Xray config has no valid inbounds array")
	}

	changed := false
	for _, inboundValue := range inbounds {
		inbound, ok := inboundValue.(map[string]any)
		if !ok || stringValue(inbound["protocol"]) != "tunnel" {
			continue
		}
		settings, ok := inbound["settings"].(map[string]any)
		if !ok {
			return false, fmt.Errorf("Xray tunnel inbound has no valid settings object")
		}

		allowedNetwork := stringValue(settings["allowedNetwork"])
		network := stringValue(settings["network"])
		switch {
		case allowedNetwork != "" && network == "":
			settings["network"] = allowedNetwork
			changed = true
		case network != "" && allowedNetwork == "":
			settings["allowedNetwork"] = network
			changed = true
		}
	}

	if !changed {
		return false, nil
	}

	coreJSON, err := json.Marshal(coreConfig)
	if err != nil {
		return false, fmt.Errorf("marshal Xray config with TProxy compatibility: %w", err)
	}
	b.coreJSON = coreJSON
	return true, nil
}
