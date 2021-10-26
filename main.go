package main

import (
	"flag"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/wietsevenema/dns-retry/server"
)

var (
	config      = &server.Config{}
	nameservers = ""
	debug       = false
	help        = false
)

func init() {
	flag.StringVar(&config.BindAddr, "listen", env("BIND_ADDR", "127.0.0.1:53"), "Adress to bind to")
	flag.StringVar(&nameservers, "nameservers", env("NAMESERVERS", ""), "List of nameservers (for example: 10.1.2.254:53,10.1.3.254:53)")
	flag.BoolVar(&debug, "debug", false, "Print debug output")
	flag.BoolVar(&help, "help", false, "Print usage information")

	flag.DurationVar(&config.ReadTimeout, "read-timeout", 2*time.Second, "Read timeout")
	flag.DurationVar(&config.WriteTimeout, "write-timeout", 2*time.Second, "Write timeout")

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	flag.Parse()

	if help {
		flag.PrintDefaults()
		os.Exit(1)
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	if nameservers != "" {
		for _, hostPort := range strings.Split(nameservers, ",") {
			config.Nameservers = append(config.Nameservers, hostPort)
		}
	}

	if err := server.SetDefaults(config); err != nil {
		log.Fatal().Err(err).Msg("Failed to set configuration")
	}

	log.Info().Str("-listen", config.BindAddr).Str("-nameservers", strings.Join(config.Nameservers, ",")).Msg("Running with: ")
	log.Debug().Msg("Debug logging enabled")

	s := server.New(config)

	if err := s.Run(); err != nil {
		log.Fatal().Err(err).Msg("Fatal error occurred")
	}
}

func env(key, def string) string {
	if x := os.Getenv(key); x != "" {
		return x
	}
	return def
}
