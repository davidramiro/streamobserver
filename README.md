# streamobserver

## About

Go service to poll Twitch and Restreamer streams and notify Telegram chats and groups.

## Setup

- Get a Telegram bot token from [here](https://t.me/BotFather)
- Get a twitch client ID and secret by registering an application [here](https://dev.twitch.tv/console/apps)
- Rename `config.sample.yml` to `config.yml` and enter your credentials
- Rename `streams.sample.yml` to `streams.yml` and enter chats to notify and streams to observe
- Either run via executable or `go run .\cmd\streamobserver\main.go` (pass `-debug` to enable lowest log level)

