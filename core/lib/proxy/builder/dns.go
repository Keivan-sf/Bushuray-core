package builder

import (
	"encoding/json"

	"bushuray-core/lib/config"
)

type xrayDNSConfig struct {
	Servers       []string                `json:"servers"`
	QueryStrategy config.DNSQueryStrategy `json:"queryStrategy"`
}

func (b *Builder) ApplyDNS(dnsConfig config.DNSConfig) error {
	var coreConfig map[string]json.RawMessage
	if err := json.Unmarshal(b.coreJSON, &coreConfig); err != nil {
		return err
	}
	if coreConfig == nil {
		return ErrCoreIsNill
	}

	dnsJSON, err := json.Marshal(xrayDNSConfig{
		Servers:       dnsConfig.Servers,
		QueryStrategy: dnsConfig.QueryStrategy,
	})
	if err != nil {
		return err
	}

	coreConfig["dns"] = dnsJSON

	coreJSON, err := json.Marshal(coreConfig)
	if err != nil {
		return err
	}

	b.coreJSON = coreJSON
	return nil
}
