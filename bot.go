package main

import (
	"log"
	"os"
	"strings"
	"time"

	"layeh.com/gumble/gumble"
	"layeh.com/gumble/gumbleutil"
)

type Bot struct {
	connectString        string
	defaultConnectString string
	channelTree          []string
}

func runBot(bot Bot) {
	gumbleutil.Main(gumbleutil.Listener{
		Connect: func(e *gumble.ConnectEvent) {
			if len(bot.channelTree) > 0 {
				e.Client.Self.Move(e.Client.Channels.Find(bot.channelTree...))
				log.Println("Connected.")
			}
			time.Sleep(time.Second * 1)
		},
		TextMessage: func(e *gumble.TextMessageEvent) {
			if strings.Contains(e.TextMessage.Message, "connect") {
				bot.connectString = e.TextMessage.Message
				e.Sender.Send("Connect received: " + bot.connectString)
				log.Printf("Connect: %v received from %v", bot.connectString, e.Sender.Name)
			}
		},
		UserChange: func(e *gumble.UserChangeEvent) {
			if e.Type.Has(gumble.UserChangeConnected) {
				if e.User.Name != "ConnectBot" {
					if e.User.Channel.Name == e.Client.Self.Channel.Name {
						if len(bot.connectString) > 0 {
							e.User.Send(bot.connectString)
							log.Printf("%v connected.\n", e.User.Name)
						}
					}
				}

			}
			if e.Type.Has(gumble.UserChangeChannel) {
				log.Printf("%v changed channel to %v.\n", e.User.Name, e.User.Channel.Name)
				if len(e.Client.Self.Channel.Users) == 1 {
					bot.connectString = bot.defaultConnectString
				}
				if e.User.Name != "ConnectBot" {
					if e.User.Channel.Name == e.Client.Self.Channel.Name {
						if len(bot.connectString) > 0 {
							e.User.Send(bot.connectString)
						}
					}
				}
			}
			if e.Type.Has(gumble.UserChangeDisconnected) {
				log.Printf("%v disconnected.\n", e.User.Name)
				log.Printf("Users: %v", len(e.Client.Self.Channel.Users))
				if len(e.Client.Self.Channel.Users) == 1 {
					bot.connectString = bot.defaultConnectString
				}
			}
		},
	})
}

func initLogging() {
	file, err := openLogFile("./bot.log")
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(file)
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)

	log.Println("log file created")
}

func openLogFile(path string) (*os.File, error) {
	logFile, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	return logFile, nil
}

func main() {
	initLogging()
	defaultConnectString := os.Getenv("DEFAULT")
	channels := os.Getenv("CHANNELS")
	channelTree := strings.Split(channels, "/") // "Others/GC channel"
	bot := Bot{channelTree: channelTree, defaultConnectString: defaultConnectString}
	runBot(bot)

}
