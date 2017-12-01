package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/miekg/dns"

	revop "github.com/fardog/reverseoperator"
	"github.com/fardog/secureoperator/cmd"
)

var (
	listenAddress = flag.String(
		"listen", ":80", "listen address, as `[host]:port`",
	)

	logLevel = flag.String(
		"level",
		"info",
		"Log level, one of: debug, info, warn, error, fatal, panic",
	)

	dnsServers = flag.String(
		"dns-servers",
		"",
		`DNS Servers used to look up the endpoint; system default is used if absent.
        Ignored if "endpoint-ips" is set. Comma separated, e.g. "8.8.8.8,8.8.4.4:53".
        The port section is optional, and 53 will be used by default.`,
	)
)

func serve(net string) {
	log.Infof("starting %s service on %s", net, *listenAddress)

	server := &dns.Server{Addr: *listenAddress, Net: net, TsigSecret: nil}
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("Failed to setup the %s server: %s\n", net, err.Error())
		}
	}()

	// serve until exit
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Infof("shutting down %s on interrupt\n", net)
	if err := server.Shutdown(); err != nil {
		log.Errorf("got unexpected error %s", err.Error())
	}
}

func main() {
	flag.Usage = func() {
		_, exe := filepath.Split(os.Args[0])
		fmt.Fprint(os.Stderr, "A DNS-over-HTTPS server with a Google DNS-over-HTTPS compatible API.\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\n  %s [options]\n\nOptions:\n\n", exe)
		flag.PrintDefaults()
	}
	flag.Parse()

	// seed the global random number generator, used in some utilities and the
	// google provider
	rand.Seed(time.Now().UTC().UnixNano())

	// set the loglevel
	level, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Fatalf("invalid log level: %s", err.Error())
	}
	log.SetLevel(level)

	dips, err := cmd.CSVtoEndpoints(*dnsServers)
	if err != nil {
		log.Fatalf("error parsing dns-servers: %v", err)
	}

	provider, err := revop.NewDNSProvider(dips)
	if err != nil {
		log.Fatal(err)
	}
	options := &revop.HandlerOptions{}
	handler := revop.NewHandler(provider, options)

	http.HandleFunc("/resolve", handler.Handle)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))

	log.Infoln("servers exited, stopping")
}
