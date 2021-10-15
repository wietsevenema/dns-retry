package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/wietsevenema/dns-retry/server"
)

var (
	config      = &server.Config{}
	nameservers = ""
)

func init() {
	flag.StringVar(&config.BindAddr, "listen", env("BIND_ADDR", "127.0.0.1:53"), "Adress to bind to (127.0.0.1:53)")
	flag.StringVar(&nameservers, "nameservers", env("NAMESERVERS", ""), "List of nameservers 10.1.2.254:53,10.1.3.254:53")
	flag.Bool("help", false, "Print usage")

	flag.DurationVar(&config.ReadTimeout, "read-timeout", 2*time.Second, "Read timeout")
	flag.DurationVar(&config.WriteTimeout, "write-timeout", 2*time.Second, "Write timeout")
}

func main() {
	flag.Parse()

	if nameservers != "" {
		for _, hostPort := range strings.Split(nameservers, ",") {
			config.Nameservers = append(config.Nameservers, hostPort)
		}
	}

	if err := server.SetDefaults(config); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Running with: ")
	fmt.Println("-listen ",
		config.BindAddr,
		" -nameservers ",
		strings.Join(config.Nameservers, ","),
	)

	s := server.New(config)

	if err := s.Run(); err != nil {
		log.Fatalf("Fatal: %s", err)
	}
}

func env(key, def string) string {
	if x := os.Getenv(key); x != "" {
		return x
	}
	return def
}
