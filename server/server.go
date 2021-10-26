package server

import (
	"net"
	"sync"

	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

type server struct {
	group        *sync.WaitGroup
	config       *Config
	dnsUDPclient *dns.Client
	dnsTCPclient *dns.Client
}

func New(config *Config) *server {
	return &server{
		config: config,
		group:  new(sync.WaitGroup),
		dnsUDPclient: &dns.Client{
			Net:            "udp",
			ReadTimeout:    config.ReadTimeout,
			WriteTimeout:   config.ReadTimeout,
			SingleInflight: true,
		},
		dnsTCPclient: &dns.Client{
			Net:            "tcp",
			ReadTimeout:    config.ReadTimeout,
			WriteTimeout:   config.ReadTimeout,
			SingleInflight: true,
		},
	}
}

func (s *server) Run() error {
	mux := dns.NewServeMux()
	mux.Handle(".", s)

	s.group.Add(1)
	go func() {
		defer s.group.Done()
		if err := dns.ListenAndServe(s.config.BindAddr, "tcp", mux); err != nil {
			log.Fatal().Err(err)
		}
	}()
	s.group.Add(1)
	go func() {
		defer s.group.Done()
		if err := dns.ListenAndServe(s.config.BindAddr, "udp", mux); err != nil {
			log.Fatal().Err(err)
		}
	}()

	s.group.Wait()
	return nil
}

func (s *server) ServeDNS(w dns.ResponseWriter, req *dns.Msg) {

	tcp := isTCP(w)

	log.Debug().
		Bool("tcp", tcp).
		Interface("questions", req.Question).
		Msg("Starting request")

	var (
		resp *dns.Msg
		err  error
	)
	for _, nameserver := range s.config.Nameservers {
		log.Debug().
			Str("nameserver", nameserver).
			Msg("Sending request")

		resp, err = s.queryWithRetry(nameserver, tcp, req)
		if err != nil || resp.Rcode == dns.RcodeNameError {
			log.Debug().Err(err).Int("Rcode", resp.Rcode).Str("nameserver", nameserver).Msg("Request failed")
			// Try next nameserver on error or non-existent domain
			continue
		}
		break
	}
	log.Debug().
		Err(err).
		Int("Rcode", resp.Rcode).
		Interface("question", req.Question).
		Interface("answers", req.Answer).
		Msg("Final result")

	if err == nil {
		resp.Compress = true
		resp.Id = req.Id
		w.WriteMsg(resp)
	} else {
		w.WriteMsg(s.ServerFailure(req))
	}

	return
}

func (s *server) queryWithRetry(nameserver string, tcp bool, req *dns.Msg) (*dns.Msg, error) {
	var (
		resp *dns.Msg
		err  error
	)
	if tcp {
		resp, err = exchangeWithRetry(s.dnsTCPclient, req, nameserver)
	} else {
		resp, err = exchangeWithRetry(s.dnsUDPclient, req, nameserver)
	}
	if err == nil {
		resp.Compress = true
		resp.Id = req.Id
		return resp, nil
	}
	return nil, err
}

// isTCP returns true if the client is connecting over TCP.
func isTCP(w dns.ResponseWriter) bool {
	_, ok := w.RemoteAddr().(*net.TCPAddr)
	return ok
}

func exchangeWithRetry(c *dns.Client, m *dns.Msg, server string) (*dns.Msg, error) {
	r, _, err := c.Exchange(m, server)
	if err == nil && r.Rcode == dns.RcodeServerFailure {
		// redo the query
		r, _, err = c.Exchange(m, server)
	}
	return r, err
}

func (s *server) ServerFailure(req *dns.Msg) *dns.Msg {
	m := new(dns.Msg)
	m.SetRcode(req, dns.RcodeServerFailure)
	return m
}
