package main

import (
	"flag"
	"fmt"
	"os"
	"streamobserver/internal/config"
	"streamobserver/internal/logger"
	"streamobserver/internal/notifier"
	"streamobserver/internal/telegram"
	"time"

	"github.com/rs/zerolog"
)

func main() {
	// get debug flag, initialize logging
	debug := flag.Bool("debug", false, "sets log level to debug")
	flag.Parse()
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	confExists, err := config.CheckPresent()
	if err != nil || !confExists {
		fmt.Fprint(os.Stdout, "ERROR: config.yml and/or streams.yml not found. Press [ENTER] to exit.")
		fmt.Scanln()
		return
	}

	config, err := config.GetConfig()
	if err != nil {
		logger.Log.Panic().Err(err)
		return
	}

	logger.InitLog(config.General.JsonLogging)

	// start telegram bot
	telegram.InitBot(*debug)

	// set up observers from streams config
	notifier.PopulateObservers()

	// set up polling ticker
	ticker := time.NewTicker(time.Duration(config.General.PollingInterval) * time.Second)
	for range ticker.C {
		notifier.Notify()
	}

}
