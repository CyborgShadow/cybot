package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/textproto"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

// Structures
// Here we identify our classifying structures

type tomlConfig struct {
	Credentials credentials
}

type credentials struct {
	User    string
	Pass    string
	Channel string
}

type Bot struct {
	server        string
	port          string
	nick          string
	user          string
	channel       string
	pass          string
	pread, pwrite chan string
	conn          net.Conn
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

func NewBot() *Bot {
	return &Bot{server: "irc.chat.twitch.tv",
		port:    "6667",
		nick:    "",
		channel: "",
		pass:    "",
		conn:    nil,
		user:    ""}
}

func (bot *Bot) Connect() (conn net.Conn, err error) {
	conn, err = net.Dial("tcp", bot.server+":"+bot.port)
	if err != nil {
		log.Fatal("unable to connect to IRC server ", err)
	}
	bot.conn = conn
	log.Printf("Connected to IRC server %s (%s)\n", bot.server, bot.conn.RemoteAddr())
	return bot.conn, nil
}

func main() {
	var config = ReadConfig()
	fmt.Printf("%+v\n", config)

	ircbot := NewBot()
	conn, _ := ircbot.Connect()
	fmt.Println("Sending password")
	conn.Write([]byte("PASS " + config.Credentials.Pass + "\r\n"))
	conn.Write([]byte("NICK " + config.Credentials.User + "\r\n"))
	fmt.Println("Joining Channel")
	conn.Write([]byte("JOIN #" + config.Credentials.Channel + "\r\n"))
	defer conn.Close()

	tp := textproto.NewReader(bufio.NewReader(conn))

	// listens/responds to chat messages
	for {
		msg, err := tp.ReadLine()
		if err != nil {
			panic(err)
		}

		// split the msg by spaces
		msgParts := strings.Split(msg, " ")

		var userInfo, action, channel, messageText string
		userInfo, msgParts = msgParts[0], msgParts[1:]
		action, msgParts = msgParts[0], msgParts[1:]
		channel, msgParts = msgParts[0], msgParts[1:]

		if len(msgParts) > 0 {
			messageText = strings.Join(msgParts, " ")
			messageText = strings.TrimLeft(messageText, ":")
		}

		log.Printf("%+v\n", msg)

		// if the msg contains PING you're required to
		// respond with PONG else you get kicked
		if userInfo == "PING" {
			conn.Write([]byte("PONG " + action))
			continue
		}

		// if msg contains PRIVMSG then respond
		if action == "PRIVMSG" {
			// echo back the same message
			conn.Write([]byte("PRIVMSG " + channel + " :" + messageText + "\r\n"))
		}
	}
}
