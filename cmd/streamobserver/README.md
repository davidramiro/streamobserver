# streamobserver

## About

Go service to poll Twitch and Restreamer streams and notify Telegram chats and groups.

## Setup

- Get a Telegram bot token from [https://t.me/BotFather](here)
- Get a twitch client ID and secret by registering an application [https://dev.twitch.tv/console/apps](here)
- Rename `config.sample.yml` to `config.yml` and enter your credentials
- Rename `streams.sample.yml` to `streams.yml` and enter chats to notify and streams to observe
- Either run via executable or `go run .\cmd\streamobserver\main.go` (pass `-debug` to enable lowest log level)

