package tunmode

var json_config_template = `{
  "log": {
    "level": "info"
  },
  "dns": {
    "fakeip": {
      "enabled": true,
      "inet4_range": "198.18.0.0/15",
      "inet6_range": "fc00::/18"
    },
    "servers": [
      {
        "tag": "dns-remote",
        "address": "8.8.8.8",
        "detour": "singbox-socks"
      },
      {
        "tag": "dns-direct",
        "address": "underlying://0.0.0.0",
        "detour": "direct"
      },
      {
        "address": "fakeip",
        "tag": "dns-fake"
      },
      {
        "address": "rcode://success",
        "tag": "dns-block"
      }
    ],
    "rules": [
      {
        "query_type": [32, 33],
        "server": "dns-block"
      },
      {
        "domain_suffix": [".lan"],
        "server": "dns-block"
      },
      {
        "process_name": ["nekobox_core", "nekobox_core.exe"],
        "server": "dns-direct"
      },
      {
        "inbound": "empty",
        "server": "dns-fake"
      }
    ]
  },
  "inbounds": [
    {
      "type": "tun",
      "tag": "tun-in",
      "interface_name": "singboxtun",
      "inet4_address": "172.19.0.1/28",

      "mtu": 9000,
      "auto_route": true,
      "strict_route": false,
      "stack": "gvisor",
      "endpoint_independent_nat": true,
      "sniff": false
    }
  ],
  "outbounds": [
    {
      "type": "socks",
      "tag": "singbox-socks",
      "udp_fragment": true,

      "server": "127.0.0.1",
      "server_port": %SOCKSPORT% 
    },
    {
      "type": "block",
      "tag": "block"
    },
    {
      "type": "direct",
      "tag": "direct"
    },
    {
      "type": "dns",
      "tag": "dns-out"
    }
  ],
  "route": {
    "final": "singbox-socks",
    "auto_detect_interface": true,
    "rules": [
      {
        "network": "udp",
        "port": [135, 137, 138, 139, 5353],
        "outbound": "block"
      },
      {
        "ip_cidr": ["224.0.0.0/3", "ff00::/8"],
        "outbound": "block"
      },
      {
        "source_ip_cidr": ["224.0.0.0/3", "ff00::/8"],
        "outbound": "block"
      },
      {
        "port": 53,
        "process_name": [
          "nekoray_core",
          "nekoray_core.exe",
          "xray",
          ".xray-wrapped"
        ],
        "outbound": "dns-out"
      },
      {
        "process_name": [
          "nekobox_core",
          "nekobox_core.exe",
          "xray",
          ".xray-wrapped"
        ],
        "outbound": "direct"
      },
      {
        "port": 53,
        "outbound": "dns-out"
      }
    ]
  }
}`
