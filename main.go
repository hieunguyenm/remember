package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hieunguyenm/remember/bot"
)

var (
	token    = flag.String("token", "", "Discord bot token")
	jsonPath = flag.String("json", "", "Path to reminders JSON")
)

func main() {
	flag.Parse()

	// Check if arguments are empty.
	if *token == "" {
		tokenEnv := os.Getenv("TOKEN")
		if tokenEnv == "" {
			log.Fatalln("Please specify token with -token <token string> or via the TOKEN environment variable")
		}
		*token = tokenEnv
	}
	if *jsonPath == "" {
		log.Fatalln("Please specify reminders JSON path with -json <path>")
	}

	// Start bot session.
	s, err := bot.NewBot(*token, *jsonPath)
	if err != nil {
		log.Fatalf("failed to make new bot: %v\n", err)
	}

	// Open connection with Discord.
	if err := s.Open(); err != nil {
		log.Fatalf("failed to open connection with Discord: %v\n", err)
	}
	log.Println("opened connection with Discord")

	// Run until interrupt.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sig
	s.Close()
}
