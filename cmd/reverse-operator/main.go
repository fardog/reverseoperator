package main

import (
	"context"
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

	shutdownTimeout = flag.Int(
		"timeout", 10, "time in seconds to hold shutdown for connected clients",
	)

	dnsServers = flag.String(
		"dns-servers",
		"8.8.8.8,8.8.4.4",
		`DNS Servers used to look up the endpoint; Comma separated, e.g.
        "8.8.8.8,8.8.4.4:53". The port section is optional, and 53 will be used
        by default.`,
	)
)

func serve(server *http.Server) {
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	// serve until exit
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Infoln("shutting down on interrupt")
	timeout := time.Duration(*shutdownTimeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Errorf("got unexpected error %s", err.Error())
	}

	<-ctx.Done()
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

	mux := http.NewServeMux()
	mux.HandleFunc("/resolve", handler.Handle)
	server := &http.Server{
		Addr:    *listenAddress,
		Handler: mux,
	}

	// start the servers
	servers := make(chan bool)
	go func() {
		serve(server)
		servers <- true
	}()

	log.Infof("server started on %v", *listenAddress)
	<-servers
	log.Infoln("servers exited, stopping")

}
