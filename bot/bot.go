package bot

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/hieunguyenm/remember/bot/handlers"
	"github.com/hieunguyenm/remember/scheduler"
)

var sch *scheduler.Scheduler

// NewBot creates a new Discord bot session.
func NewBot(token string, jsonPath string) (*discordgo.Session, error) {
	s, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("failed to make new bot session: %v", err)
	}
	s.AddHandler(handle)

	// Make scheduler and read reminder JSON if exists and assign it.
	c := scheduler.NewScheduler(jsonPath)
	if err := c.LoadJSON(s, jsonPath); err != nil {
		return nil, fmt.Errorf("failed to load reminder JSON: %v", err)
	}
	sch = c
	return s, nil
}

func handle(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages sent by itself.
	if m.Author.ID == s.State.User.ID {
		return
	}
	switch {
	case m.Content == "!ping":
		go s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+" pong!")
	case strings.HasPrefix(m.Content, "!newreminder "):
		go handlers.NewReminder(s, m, sch, strings.TrimPrefix(m.Content, "!newreminder "))
	case strings.HasPrefix(m.Content, "!delreminder "):
		go handlers.DeleteReminder(s, m, sch, strings.TrimPrefix(m.Content, "!delreminder "))
	case m.Content == "!lsreminder":
		go handlers.ListReminders(s, m, sch)
	}
}
