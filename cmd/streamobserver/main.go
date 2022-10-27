package main

import (
	"flag"
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
	logger.InitLog(*debug)

	// start telegram bot
	telegram.InitBot()

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
