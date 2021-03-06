package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/astravexton/irchuu/config"
	irchuubase "github.com/astravexton/irchuu/db"

	// "github.com/astravexton/irchuu/hq" // we don't need this
	irchuu "github.com/astravexton/irchuu/irc"
	"github.com/astravexton/irchuu/paths"
	"github.com/astravexton/irchuu/relay"
	mediaserver "github.com/astravexton/irchuu/server"
	"github.com/astravexton/irchuu/telegram"
)

func main() {
	fmt.Printf("IRChuu! v%v (https://github.com/astravexton/irchuu)\n", config.VERSION)

	configFile, dataDir := paths.GetPaths()

	flag.StringVar(&configFile, "config", configFile, "path to the configuration file (will be created if not exists)")
	flag.StringVar(&dataDir, "data", dataDir, "path to the data dir")

	flag.Parse()

	err := paths.MakePaths(configFile, dataDir)
	if err != nil {
		os.Exit(1)
	}

	log.Printf("Using configuration file: %v\n", configFile)
	log.Printf("Using data directory: %v\n", dataDir)
	err, irc, tg, irchuuConf := config.ReadConfig(configFile)
	if err != nil {
		log.Fatalf("Unable to parse the config: %v\n", err)
	}

	r := relay.NewRelay()

	if irchuuConf.DBURI != "" {
		irchuubase.Init(irchuuConf.DBURI)
	}

	tg.DataDir = dataDir

	if tg.Storage == "server" {
		go mediaserver.Serve(tg)
	}

	// hq.Report(irchuuConf, tg, irc)

	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go sigNotify(sigCh, r)

	var wg sync.WaitGroup
	wg.Add(2)
	go irchuu.Launch(irc, &wg, r)
	go telegram.Launch(tg, &wg, r)
	wg.Wait()
}

func sigNotify(sigCh chan os.Signal, r *relay.Relay) {
	sig := <-sigCh
	log.Printf("Caught signal: %v, exiting...\n", sig)
	r.TeleServiceCh <- relay.ServiceMessage{Command: "shutdown", Arguments: []string{}}
}
