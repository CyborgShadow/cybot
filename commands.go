package main

import (
	"strings"
	"time"
)

// Here's our sample trigger.
// Copy this and use as a template whenever adding new commands.
var sampleTrigger = Trigger{
	func(bot *Bot, m *Message) bool {
		return m.Command == "PRIVMSG" && m.Content == "!sample"
	},
	func(irc *Bot, m *Message) bool {
		irc.Reply(m, "sample")
		return false
	},
}

// This trigger is a sample long message with a sleep in it.
var LongTrigger = Trigger{
	func(bot *Bot, m *Message) bool {
		return m.Command == "PRIVMSG" && m.Content == "-long"
	},
	func(irc *Bot, m *Message) bool {
		irc.Reply(m, "This is the first message")
		time.Sleep(5 * time.Second)
		irc.Reply(m, "This is the second message")

		return false
	},
}

// This makes the bot say hello when you say hello.
var SayHello = Trigger{
	func(bot *Bot, m *Message) bool {
		return m.Command == "PRIVMSG" && strings.ToUpper(m.Trailing) == "HELLO"
	},
	func(irc *Bot, m *Message) bool {
		irc.Reply(m, "Hello "+m.From+" and welcome to the home of cyborgshadow!")
		return false
	},
}

// This command returns and answers !uptime with bot uptime
var Uptime = Trigger{
	func(bot *Bot, m *Message) bool {
		return m.Command == "PRIVMSG" && m.Content == "!uptime"
	},
	func(irc *Bot, m *Message) bool {
		irc.Reply(m, "I've been awake for "+irc.Uptime()+" seconds, please let me sleep...")
		return false
	},
}
