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
	connect_string string
	channel        string
	subchannel     bool
	channel_ID     int32
}

func runBot(bot Bot) {
	gumbleutil.Main(gumbleutil.Listener{
		Connect: func(e *gumble.ConnectEvent) {
			if !bot.subchannel {
				log.Println("HL")
				e.Client.Self.Move(e.Client.Channels.Find(bot.channel))
			} else {
				e.Client.Self.Move(e.Client.Channels.Find(bot.channel, "GC channel"))
			}
			log.Println("Connected.")
		},
		TextMessage: func(e *gumble.TextMessageEvent) {
			if strings.Contains(e.TextMessage.Message, "connect") {
				bot.connect_string = e.TextMessage.Message
				e.Sender.Send("Connect received: " + bot.connect_string)
				log.Printf("Connect: %v received from %v", bot.connect_string, e.Sender.Name)
			}
		},
		UserChange: func(e *gumble.UserChangeEvent) {
			if e.Type.Has(gumble.UserChangeConnected) {
				if !bot.subchannel {
					e.Client.Self.Move(e.Client.Channels.Find(bot.channel))
				} else {
					if e.User.Channel == e.Client.Channels.Find(bot.channel, "GC channel") {
						log.Printf("%v connected.\n", e.User.Name)
						e.User.Send(bot.connect_string)
					}
				}

			}
			if e.Type.Has(gumble.UserChangeChannel) {
				log.Printf("%v changed channel to %v.\n", e.User.Name, e.User.Channel.Name)
				if len(e.Client.Self.Channel.Users) == 1 {
					bot.connect_string = nipple
				}
				if e.User.Name != "ConnectBot" {
					if e.User.Channel.ID == uint32(bot.channel_ID) {
						e.User.Send(bot.connect_string)
					}
				}
			}
			if e.Type.Has(gumble.UserChangeDisconnected) {
				log.Printf("%v disconnected.\n", e.User.Name)
				log.Printf("Users: %v", len(e.Client.Self.Channel.Users))
				if len(e.Client.Self.Channel.Users) == 1 {
					bot.connect_string = nipple
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

	sixes := Bot{connect_string: nipple, channel: "Others", subchannel: true, channel_ID: 1384}
	hl := Bot{connect_string: nipple, channel: "9v9 Xenon", subchannel: false, channel_ID: 18}

	if strings.Contains(os.Args[1], "icewind") {
		runBot(hl)
	} else {
		runBot(sixes)
	}
}
