package main

import (
	"flag"
	"fmt"
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
	logger.InitLog()

	confExists, err := config.CheckPresent()
	if err != nil || !confExists {
		logger.Log.Error().Msg("config.yml and/or streams.yml not found. Press [ENTER] to exit.")
		fmt.Scanln()
		return
	}

	// start telegram bot
	telegram.InitBot(*debug)

	// set up observers from streams config
	notifier.PopulateObservers()

	// set up polling ticker
	config, err := config.GetConfig()
	if err != nil {
		logger.Log.Panic().Err(err)
		return
	}
	ticker := time.NewTicker(time.Duration(config.General.PollingInterval) * time.Second)
	for range ticker.C {
		notifier.Notify()
	}

}
