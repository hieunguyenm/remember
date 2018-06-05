package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/robfig/cron"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var token string
var cronRunner = cron.New()
var tempStore = map[string]Reminder{}
var runningReminders []Reminder

func main() {
	if len(os.Args) < 2 {
		log.Fatalln("Discord bot token not supplied as argument. Try again with: ./remember <token>")
	}

	token = os.Args[1]

	// Create a new Discord session using the provided bot token.
	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Error creating Discord session : %v", err)
		return
	}

	ReadJSON()
	log.Println("Reminders JSON loaded.")

	cronRunner.Start()
	discord.AddHandler(chooseHandler)

	// Open a websocket connection to Discord and begin listening.
	err = discord.Open()
	if err != nil {
		log.Printf("Error establishing connection: %v", err)
		return
	}

	log.Println("Ready to remember...")

	// Catch interrupts
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	discord.Close()
}

func chooseHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	switch {
	case m.Content == "!ping":
		s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+" pong!")
	case m.Content == "!start":
		startReminders(s, m)
	case m.Content == "!help":
		sendHelp(s, m)
	case m.Content == "!clearnew":
		clearIncomplete(s, m)
	case m.Content == "!showreminders":
		showAll(s, m)
	case strings.HasPrefix(m.Content, "!stopremind"):
		stopReminder(s, m)
	case strings.HasPrefix(m.Content, "!newremind"):
		addNew(s, m)
	case strings.HasPrefix(m.Content, "!setdate"):
		setDate(s, m)
	case strings.HasPrefix(m.Content, "!setinterval"):
		setInterval(s, m)
	case strings.HasPrefix(m.Content, "!setdescription"):
		setDescription(s, m)
	}
}
