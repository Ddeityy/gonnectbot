package main

import (
	"log"
	"os"
	"strings"

	"layeh.com/gumble/gumble"
	"layeh.com/gumble/gumbleutil"
)

const nipple = "connect nipple.tf; password nipple"

type Bot struct {
	connectString    string
	channel          string
	subchannel       bool
	subchannelString string
	channelId        int32
}

func runBot(bot Bot) {
	gumbleutil.Main(gumbleutil.Listener{
		Connect: func(e *gumble.ConnectEvent) {
			if !bot.subchannel {
				e.Client.Self.Move(e.Client.Channels.Find(bot.channel))
			} else {
				e.Client.Self.Move(e.Client.Channels.Find(bot.channel, bot.subchannelString))
			}
			log.Println("Connected.")
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
					if e.User.Channel.ID == uint32(bot.channelId) {
						log.Printf("%v connected.\n", e.User.Name)
					}
				}

			}
			if e.Type.Has(gumble.UserChangeChannel) {
				log.Printf("%v changed channel to %v.\n", e.User.Name, e.User.Channel.Name)
				if len(e.Client.Self.Channel.Users) == 1 {
					bot.connectString = nipple
				}
				if e.User.Name != "ConnectBot" {
					if e.User.Channel.ID == uint32(bot.channelId) {
						e.User.Send(bot.connectString)
					}
				}
			}
			if e.Type.Has(gumble.UserChangeDisconnected) {
				log.Printf("%v disconnected.\n", e.User.Name)
				log.Printf("Users: %v", len(e.Client.Self.Channel.Users))
				if len(e.Client.Self.Channel.Users) == 1 {
					bot.connectString = nipple
				}
			}
		},
	})
}

func main() {
	fileName := "bot.log"
	logFile, err := os.OpenFile(fileName, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Panic(err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	sixes := Bot{connectString: nipple, channel: "Others", subchannelString: "GC channel", subchannel: true, channelId: 1384}
	hl := Bot{connectString: nipple, channel: "9v9 Xenon", subchannelString: "", subchannel: false, channelId: 18}

	if strings.Contains(os.Args[1], "icewind") {
		runBot(hl)
	} else {
		runBot(sixes)
	}
}
