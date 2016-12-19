package main

import (
	"strings"
	"time"
)

// Here's our sample trigger.
// Copy this and use as a template whenever adding new commands.
var sampleTrigger = Trigger{
	func(bot *Bot, tm *TwitchMessage) bool {
		return tm.Message.Command == "PRIVMSG" && tm.Message.Content == "!sample"
	},
	func(irc *Bot, tm *TwitchMessage) bool {
		irc.Reply(tm, "sample")
		return false
	},
}

// Here's the cheat sheet of all the command parameters you can access
// The first Block are specific to Twitch messages
// The rest are generic IRC and happen on every message in Twitch
//tm.Badges -- What badges the user has on, "Broadcaster" and either of "turbo and premium (prime)"
//tm.Color -- What their chat color was
//tm.DisplayName -- The user's display name
//tm.Emotes -- What emotes were used in their message
//tm.Id -- No idea??
//tm.Mod -- Whether or not the user is a mod (Note: is false for channel owners)
//tm.RoomId -- The twitch room ID
//tm.SentTs -- The timestamp of the message
//tm.Turbo -- Whether or not the user is Twitch Turbo
//tm.UserId -- The Numerical UserId of the sender
//tm.UserType -- The User Type -- almost always useless
//
//
//tm.Message.To -- Who the message was sent to
//tm.Message.From -- Who the message is from
//tm.Message.Trailing -- What trails the message (often useless)
//tm.Message.Content -- The content of the message sent (Most common thing you'll use)
//tm.Message.Prefix.Name -- the URI used by the message (almost never used)
//tm.Message.Prefix.User -- Blank on Twitch
//tm.Message.Prefix.Host -- Blank on Twitch
//tm.Message.Command -- IRC Commands, Ones you'll be interested in are ACTION (/me, etc.) and PRIVMSG (normal messages)
//tm.Message.Params -- Extra params passed along (often the channel name)

// This trigger is a sample long message with a sleep in it.
var LongTrigger = Trigger{
	func(bot *Bot, tm *TwitchMessage) bool {
		return tm.Message.Command == "PRIVMSG" && strings.ToUpper(tm.Message.Content) == "-LONG"
	},
	func(irc *Bot, tm *TwitchMessage) bool {
		irc.Reply(tm, "This is the first message")
		time.Sleep(5 * time.Second)
		irc.Reply(tm, "This is the second message")

		return false
	},
}

// This makes the bot say hello when you say hello.
var SayHello = Trigger{
	func(bot *Bot, tm *TwitchMessage) bool {
		return tm.Message.Command == "PRIVMSG" && strings.ToUpper(tm.Message.Content) == "HELLO"
	},
	func(irc *Bot, tm *TwitchMessage) bool {
		irc.Reply(tm, "Hello "+tm.DisplayName+" and welcome to "+strings.TrimLeft(tm.Message.Params[0], "#")+"'s chat room.")
		return false
	},
}

// This command returns and answers !uptime with bot uptime
var Uptime = Trigger{
	func(bot *Bot, tm *TwitchMessage) bool {
		return tm.Message.Command == "PRIVMSG" && strings.ToUpper(tm.Message.Content) == "!UPTIME"
	},
	func(irc *Bot, tm *TwitchMessage) bool {
		irc.Reply(tm, "I've been awake for "+irc.Uptime()+" seconds, please let me sleep...")
		return false
	},
}
