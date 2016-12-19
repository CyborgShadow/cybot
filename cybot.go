package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

type tomlConfig struct {
	Credentials credentials
}

type credentials struct {
	User     string
	Pass     string
	Channels []string
}

// Reads info from config file
func ReadConfig() tomlConfig {
	var configfile = "twitch_credentials.toml"
	_, err := os.Stat(configfile)
	if err != nil {
		log.Fatal("Config file is missing: ", configfile)
	}

	var config tomlConfig
	if _, err := toml.DecodeFile(configfile, &config); err != nil {
		log.Fatal(err)
	}

	// log.Printf("%+v\n", config)
	return config
}

var serv = flag.String("server", "irc.chat.twitch.tv:6667", "hostname and port for irc server to connect to")
var config = ReadConfig()
var nick = flag.String("nick", config.Credentials.User, "nickname for the bot")
var pass = flag.String("pass", config.Credentials.Pass, "password for the bot")

func main() {
	flag.Parse()

	hijackSession := func(bot *Bot) {
		bot.HijackSession = true
	}
	channels := func(bot *Bot) {
		bot.Channels = config.Credentials.Channels
	}
	irc, err := NewBot(*serv, *nick, *pass, hijackSession, channels)
	if err != nil {
		panic(err)
	}

	irc.AddTrigger(LongTrigger)
	irc.AddTrigger(SayHello)
	irc.AddTrigger(Uptime)

	// Start up bot (this blocks until we disconnect)
	irc.Run()
	fmt.Println("Bot shutting down.")
}
