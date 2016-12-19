package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/inconshreveable/log15"
	logext "github.com/inconshreveable/log15/ext"
)

// Bot implements an irc bot to be connected to a given server
type Bot struct {

	// This is set if we have hijacked a connection
	reconnecting bool
	// Channel for user to read incoming messages
	Incoming chan *Message
	con      net.Conn
	outgoing chan string
	triggers []Trigger
	// When did we start? Used for uptime
	started time.Time
	// Unix domain socket address for reconnects (linux only)
	unixastr string
	unixlist net.Listener
	// Log15 loggger
	log.Logger
	didJoinChannels sync.Once

	// Exported fields
	Host          string
	Password      string
	Channels      []string
	HijackSession bool
	// This bots nick
	Nick string
	// Duration to wait between sending of messages to avoid being
	// kicked by the server for flooding (default 200ms)
	ThrottleDelay time.Duration
	// Maxmimum time between incoming data
	PingTimeout time.Duration

	TLSConfig tls.Config
}

func (bot *Bot) String() string {
	return fmt.Sprintf("Server: %s, Channels: %v, Nick: %s", bot.Host, bot.Channels, bot.Nick)
}

// NewBot creates a new instance of Bot
func NewBot(host, nick string, pass string, options ...func(*Bot)) (*Bot, error) {
	// Defaults are set here
	bot := Bot{
		Incoming:      make(chan *Message, 16),
		outgoing:      make(chan string, 16),
		started:       time.Now(),
		unixastr:      fmt.Sprintf("@%s-%s/bot", host, nick),
		Host:          host,
		Nick:          nick,
		ThrottleDelay: 200 * time.Millisecond,
		PingTimeout:   300 * time.Second,
		HijackSession: false,
		Channels:      []string{"#test"},
		Password:      pass,
	}
	for _, option := range options {
		option(&bot)
	}
	// Discard logs by default
	bot.Logger = log.New("id", logext.RandId(8), "host", bot.Host, "nick", log.Lazy{bot.getNick})

	bot.Logger.SetHandler(log.DiscardHandler()) // Use this one for prod pushes
	bot.Logger.SetHandler(log.StdoutHandler)    // Use this one for debugging
	bot.AddTrigger(pingPong)
	bot.AddTrigger(joinChannels)
	return &bot, nil
}

// Uptime returns the uptime of the bot
func (bot *Bot) Uptime() string {
	// return fmt.Sprintf("Started: %s, Uptime: %s", bot.started, time.Since(bot.started))
	return fmt.Sprintf("%s", time.Since(bot.started))
}

func (bot *Bot) getNick() string {
	return bot.Nick
}

func (bot *Bot) connect(host string) (err error) {
	bot.Debug("Connecting")
	bot.con, err = net.Dial("tcp", host)
	return
}

// Incoming message gathering routine
func (bot *Bot) handleIncomingMessages() {
	scan := bufio.NewScanner(bot.con)
	for scan.Scan() {
		// Disconnect if we have seen absolutely nothing for 300 seconds
		bot.con.SetDeadline(time.Now().Add(bot.PingTimeout))
		//DEBUG ONLY - DELETE LATER
		bot.Debug(scan.Text())
		msg := ParseMessage(scan.Text())
		// bot.Debug("Incoming", "msg.To", msg.To, "msg.From", msg.From, "msg.Params", msg.Params, "msg.Trailing", msg.Trailing)
		for _, t := range bot.triggers {
			if t.Condition(bot, msg) {
				go t.Action(bot, msg)
			}
		}
		bot.Incoming <- msg
	}
	close(bot.Incoming)
}

// Handles message speed throtling
func (bot *Bot) handleOutgoingMessages() {
	for s := range bot.outgoing {
		bot.Debug("Outgoing", "data", s)
		_, err := fmt.Fprint(bot.con, s+"\r\n")
		if err != nil {
			bot.Error("handleOutgoingMessages fmt.Fprint error", "err", err)
			return
		}
		time.Sleep(bot.ThrottleDelay)
	}
}

// StandardRegistration performs a basic set of registration commands
func (bot *Bot) StandardRegistration() {
	//Server registration
	if bot.Password != "" {
		bot.Send("PASS " + bot.Password)
	}
	bot.Debug("Sending standard registration")
	bot.sendUserCommand(bot.Nick, bot.Nick, "8")
	bot.SetNick(bot.Nick)
	bot.Send("CAP REQ :twitch.tv/tags")
}

// Set username, real name, and mode
func (bot *Bot) sendUserCommand(user, realname, mode string) {
	bot.Send(fmt.Sprintf("USER %s %s * :%s", user, mode, realname))
}

// SetNick sets the bots nick on the irc server
func (bot *Bot) SetNick(nick string) {
	bot.Nick = nick
	bot.Send(fmt.Sprintf("NICK %s", nick))
}

// Run starts the bot and connects to the server. Blocks until we disconnect from the server.
func (bot *Bot) Run() {
	bot.Debug("Starting bot goroutines")

	// Attempt reconnection
	var hijack bool
	if bot.HijackSession {
		hijack = bot.hijackSession()
		bot.Debug("Hijack", "Did we?", hijack)
	}

	if !hijack {
		err := bot.connect(bot.Host)
		if err != nil {
			bot.Crit("bot.Connect error", "err", err.Error())
			return
		}
		bot.Info("Connected successfully!")
	}

	go bot.handleIncomingMessages()
	go bot.handleOutgoingMessages()

	go bot.StartUnixListener()

	// Only register on an initial connection
	if !bot.reconnecting {
		bot.StandardRegistration()
	}
	for m := range bot.Incoming {
		if m == nil {
			log.Info("Disconnected")
			return
		}
	}
}

// Reply sends a message to where the message came from (user or channel)
func (bot *Bot) Reply(m *Message, text string) {
	var target string
	if strings.Contains(m.To, "#") {
		target = m.To
	} else {
		target = m.From
	}
	bot.Msg(target, text)
}

// Msg sends a message to 'who' (user or channel)
func (bot *Bot) Msg(who, text string) {
	for len(text) > 400 {
		bot.Send("PRIVMSG " + who + " :" + text[:400])
		text = text[400:]
	}
	bot.Send("PRIVMSG " + who + " :" + text)
}

// Notice sends a NOTICE message to 'who' (user or channel)
func (bot *Bot) Notice(who, text string) {
	for len(text) > 400 {
		bot.Send("NOTICE " + who + " :" + text[:400])
		text = text[400:]
	}
	bot.Send("NOTICE " + who + " :" + text)
}

// Action sends an action to 'who' (user or channel)
func (bot *Bot) Action(who, text string) {
	msg := fmt.Sprintf("\u0001ACTION %s\u0001", text)
	bot.Msg(who, msg)
}

// Topic sets the channel 'c' topic (requires bot has proper permissions)
func (bot *Bot) Topic(c, topic string) {
	str := fmt.Sprintf("TOPIC %s :%s", c, topic)
	bot.Send(str)
}

// Send any command to the server
func (bot *Bot) Send(command string) {
	bot.outgoing <- command
}

// ChMode is used to change users modes in a channel
// operator = "+o" deop = "-o"
// ban = "+b"
func (bot *Bot) ChMode(user, channel, mode string) {
	bot.Send("MODE " + channel + " " + mode + " " + user)
}

// Join a channel
func (bot *Bot) Join(ch string) {
	bot.Send("JOIN " + ch)
}

// Close closes the bot
func (bot *Bot) Close() error {
	if bot.unixlist != nil {
		return bot.unixlist.Close()
	}
	return nil
}

// AddTrigger adds a given trigger to the bots handlers
func (bot *Bot) AddTrigger(t Trigger) {
	bot.triggers = append(bot.triggers, t)
}

// Trigger is used to subscribe and react to events on the bot Server
type Trigger struct {
	// Returns true if this trigger applies to the passed in message
	Condition func(*Bot, *Message) bool

	// The action to perform if Condition is true
	// return true if the message was 'consumed'
	Action func(*Bot, *Message) bool
}

// A trigger to respond to the servers ping pong messages
// If PingPong messages are not responded to, the server assumes the
// client has timed out and will close the connection.
// Note: this is automatically added in the IrcCon constructor
var pingPong = Trigger{
	func(bot *Bot, m *Message) bool {
		return m.Command == "PING"
	},
	func(bot *Bot, m *Message) bool {
		fmt.Println("Return message is: PONG :" + m.Content)
		bot.Send("PONG :" + m.Content)
		return true
	},
}

var joinChannels = Trigger{
	func(bot *Bot, m *Message) bool {
		return m.Command == strconv.Itoa(001) || m.Command == strconv.Itoa(376) // 001 and 376 are irc's RPL_WELCOME and RPL_ENDMOTD
	},
	func(bot *Bot, m *Message) bool {
		bot.didJoinChannels.Do(func() {
			for _, channel := range bot.Channels {
				splitchan := strings.SplitN(channel, ":", 2)
				fmt.Println("splitchan is:", splitchan)
				if len(splitchan) == 2 {
					channel = splitchan[0]
					password := splitchan[1]
					bot.Send(fmt.Sprintf("JOIN %s %s", channel, password))
				} else {
					bot.Send(fmt.Sprintf("JOIN %s", channel))
				}
			}
		})
		return true
	},
}

func ReconOpt() func(*Bot) {
	return func(b *Bot) {
		b.HijackSession = true
	}
}

// Message represents a message received from the server
type Message struct {
	// Content generally refers to the text of a PRIVMSG
	Content string

	// Command generated from the server
	Command string

	Params []string

	Name string // Nick- or servername

	User string //Username

	Host     string // Hostname
	Trailing string
	//Time at which this message was recieved
	TimeStamp time.Time

	// Entity that this message was addressed to (channel or user)
	To string

	// Nick of the messages sender (equivalent to Prefix.Name)
	// Outdated, please use .Name
	From string
}

// ParseMessage takes a string and attempts to create a Message struct.
// Returns nil if the Message is invalid.
func ParseMessage(raw string) (m *Message) {
	m = new(Message)

	// Cut up our message into the 3 slices - Command, Params, Trailing
	var mSplice []string
	cutRaw := raw

	if strings.Count(cutRaw, ":") >= 2 {
		for i := 0; i < 3; i++ {
			fmt.Println("cutRaw is : " + cutRaw)

			if i < 2 {
				fmt.Println("Colon index is :" + strconv.Itoa(strings.Index(cutRaw, ":")))
				fmt.Println("The string you want is: " + cutRaw[:strings.Index(cutRaw, ":")])
				mSplice = append(mSplice, cutRaw[:strings.Index(cutRaw, ":")])
				cutRaw = cutRaw[strings.Index(cutRaw, ":")+1:]

			} else if i == 2 {
				mSplice = append(mSplice, cutRaw)
			}
			fmt.Println("mSplice value is : " + mSplice[i])
		}
	}

	//DEBUG ONLY - REMOVE THESE
	fmt.Println("The raw message is: " + raw)
	fmt.Println("The number of Splices is: " + strconv.Itoa(len(mSplice)))
	fmt.Println("The slices are: %+v", mSplice)
	fmt.Println("Splice part 1 is: " + mSplice[0])
	fmt.Println("Splice part 2 is: " + mSplice[1])
	fmt.Println("Splice part 3 is: " + mSplice[2])

	commandSplice := strings.Split(mSplice[1], " ")

	fmt.Println("commandSplcies are: %+v", commandSplice)
	fmt.Println("Splice part 1 is: " + commandSplice[0])
	fmt.Println("Splice part 2 is: " + commandSplice[1])
	fmt.Println("Splice part 3 is: " + commandSplice[2])

	if len(mSplice) == 3 {
		m.From = commandSplice[0]
		m.To = commandSplice[2]
		m.Command = commandSplice[1]
		m.Content = mSplice[2]

	} else if len(mSplice) == 4 {
		m.From = mSplice[0]
		// m.Command = mSplice[1]
		m.To = mSplice[2]
		m.Content = mSplice[3]
	}
	//	if len(mSplice.Params) > 0 {
	//		m.From = strings.TrimLeft(m.Params[0], ':')
	//		m.To = m.Params[2]
	//	}
	// 	m.Content = m.Trailing

	// DEBUG ONLY - REMOVE THESE
	//fmt.Println("Name is: " + m.Message2.Prefix.Name)
	fmt.Println("Content is: " + m.Content)
	fmt.Println("To is: " + m.To)
	fmt.Println("From is: " + m.From)
	//	if len(m.Params) > 0 {
	//		m.To = m.Params[0]
	//	} else if m.Command == "JOIN" {
	//		m.To = m.Trailing
	//	}
	//if m.Prefix != nil {
	//	m.From = m.Prefix.Name
	//}
	m.TimeStamp = time.Now()

	return m
}
