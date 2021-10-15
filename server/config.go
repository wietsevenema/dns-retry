package server

import (
	"net"
	"os"
	"time"

	"github.com/miekg/dns"
)

type Config struct {
	BindAddr     string        `json:"bind_addr,omitempty"`
	ReadTimeout  time.Duration `json:"read_timeout,omitempty"`
	WriteTimeout time.Duration `json:"write_timeout,omitempty"`
	Nameservers  []string      `json:"nameservers,omitempty"`
}

func SetDefaults(config *Config) error {
	if config.ReadTimeout == 0 {
		config.ReadTimeout = 1 * time.Second
	}
	if config.WriteTimeout == 0 {
		config.WriteTimeout = 1 * time.Second
	}
	if len(config.Nameservers) == 0 {
		c, err := dns.ClientConfigFromFile("/etc/resolv.conf")
		if !os.IsNotExist(err) {
			if err != nil {
				return err
			}
			for _, s := range c.Servers {
				config.Nameservers = append(config.Nameservers, net.JoinHostPort(s, c.Port))
			}
		}
	}
	return nil
}
