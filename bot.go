package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"layeh.com/gumble/gumble"
	"layeh.com/gumble/gumbleffmpeg"
	"layeh.com/gumble/gumbleutil"

	_ "layeh.com/gumble/opus"
)

type Bot struct {
	connectString        string
	defaultConnectString string
	channelTree          []string
}

func client(cchan chan []string, dchan chan string, listeners ...gumble.EventListener) {
	server := flag.String("server", "localhost:64738", "Mumble server address")
	username := flag.String("username", "gumble-bot", "client username")
	password := flag.String("password", "", "client password")
	insecure := flag.Bool("insecure", false, "skip server certificate verification")
	certificateFile := flag.String("certificate", "", "user certificate file (PEM)")
	keyFile := flag.String("key", "", "user certificate key file (PEM)")

	defaultConnectString := flag.String("default", "", "default string to send out")
	channels := flag.String("channel", "", "channel names separated by `/` `root/channel/subchannel`")

	if !flag.Parsed() {
		flag.Parse()
	}

	channelTree := strings.Split(*channels, "/")

	go func() {
		cchan <- channelTree
	}()

	go func() {
		dchan <- *defaultConnectString
	}()

	defer close(cchan)
	defer close(dchan)

	host, port, err := net.SplitHostPort(*server)
	if err != nil {
		host = *server
		port = strconv.Itoa(gumble.DefaultPort)
	}

	keepAlive := make(chan bool)

	config := gumble.NewConfig()
	config.Username = *username
	config.Password = *password

	address := net.JoinHostPort(host, port)

	var tlsConfig tls.Config

	if *insecure {
		tlsConfig.InsecureSkipVerify = true
	}

	if *certificateFile != "" {
		if *keyFile == "" {
			keyFile = certificateFile
		}
		if certificate, err := tls.LoadX509KeyPair(*certificateFile, *keyFile); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err)
			os.Exit(1)
		} else {
			tlsConfig.Certificates = append(tlsConfig.Certificates, certificate)
		}
	}

	config.Attach(gumbleutil.AutoBitrate)
	for _, listener := range listeners {
		config.Attach(listener)
	}

	config.Attach(gumbleutil.Listener{
		Disconnect: func(e *gumble.DisconnectEvent) {
			keepAlive <- true
		},
	})

	_, err = gumble.DialWithDialer(new(net.Dialer), address, config, &tlsConfig)
	if err != nil {
		log.Printf("%s: %s\n", os.Args[0], err)
		time.Sleep(time.Second * 30)
		client(cchan, dchan, listeners...)
	}

	<-keepAlive
}

func runBot(bot Bot) {
	cchan := make(chan []string)
	dchan := make(chan string)

	client(cchan, dchan, gumbleutil.Listener{
		Connect: func(e *gumble.ConnectEvent) {
			bot.channelTree = <-cchan
			bot.defaultConnectString = <-dchan
			if len(bot.channelTree) > 0 {
				e.Client.Self.Move(e.Client.Channels.Find(bot.channelTree...))
				log.Println("Connected.")
			}
			bot.connectString = bot.defaultConnectString
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
				if e.User.Name != "_ConnectBot" {
					if e.User.Channel.Name == e.Client.Self.Channel.Name {
						log.Printf("%v connected.\n", e.User.Name)
					}
				}

			}
			if e.Type.Has(gumble.UserChangeChannel) {
				log.Printf("%v changed channel to %v.\n", e.User.Name, e.User.Channel.Name)
				if len(e.Client.Self.Channel.Users) == 1 {
					bot.connectString = bot.defaultConnectString
				}
				if e.User.Name != "_ConnectBot" {
					if e.User.Channel.Name == e.Client.Self.Channel.Name {
						if e.User.Name == "Fish" || e.User.Name == "fishage" {
							if fishLate() {
								bullyFish(e)
							}
						}
						log.Println(bot.connectString)
						if len(bot.connectString) > 0 {
							e.User.Send(bot.connectString)
							log.Printf("Sent connect to %s", e.User.Name)
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

func fishLate() bool {
	now := time.Now()

	if now.Minute() >= 30 {
		now.Local().Add(time.Hour * 1)
	}

	rounded := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, now.Location())

	return now.After(rounded)
}

func bullyFish(e *gumble.UserChangeEvent) {
	var stream *gumbleffmpeg.Stream
	if stream != nil && stream.State() == gumbleffmpeg.StatePlaying {
		return
	}

	stream = gumbleffmpeg.New(e.Client, gumbleffmpeg.SourceFile("./fish/fish1.mp3"))
	if err := stream.Play(); err != nil {
		log.Printf("%s\n", err)
	} else {
		log.Printf("Playing %s\n", "./fish/fish1.mp3")
	}
}

func main() {
	bot := Bot{}
	runBot(bot)
}
