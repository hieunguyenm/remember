package handlers

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/hieunguyenm/remember/scheduler"
	"gopkg.in/robfig/cron.v2"
)

var cronSpecNonStandard = regexp.MustCompile("@.*")

// NewReminder creates a new reminder.
func NewReminder(s *discordgo.Session, m *discordgo.MessageCreate, c *scheduler.Scheduler, msg string) {
	// Split into [cron expression, target date, reminder description].
	parts := strings.Split(msg, ";")
	if len(parts) != 3 {
		s.ChannelMessageSend(m.ChannelID,
			fmt.Sprintf("Expected 3 parameters, got %d. Please try again.", len(parts)))
		return
	}

	// Remove leading and trailing whitespaces.
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	// Check if spec is non-standard.
	var spec string
	if cronSpecNonStandard.MatchString(parts[0]) {
		spec = parts[0]
	} else {
		// Run at 10 second mark to account for delays.
		spec = "10 " + parts[0]
	}

	// Check if spec is valid.
	if _, err := cron.Parse(spec); err != nil {
		s.ChannelMessageSend(m.ChannelID,
			fmt.Sprintf("Invalid cron spec, got `%s`. See https://crontab.guru/.", parts[0]))
		return
	}

	// Check if target date is valid.
	parsedDate, err := time.Parse("02-01-2006", parts[1])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Invalid target date, got `%s`. Make sure the format is DD-MM-YYYY.", parts[1]))
		return
	}

	// Check for empty description.
	if parts[2] == "" {
		s.ChannelMessageSend(m.ChannelID, "Empty description supplied. Please try again.")
		return
	}

	// Schedule reminder.
	if err := c.ScheduleReminder(s, &scheduler.Reminder{
		Date:        parts[1],
		Description: parts[2],
		Interval:    parts[0],
		ParsedTime:  parsedDate,
		Channel:     m.ChannelID,
	}); err != nil {
		s.ChannelMessageSend(m.ChannelID, "Failed to schedule reminder.")
		log.Printf("failed to schedule reminder: %v\n", err)
		return
	}
	s.ChannelMessageSend(m.ChannelID,
		fmt.Sprintf("Reminder `%s` set for `%s` at interval `%s`.", parts[2], parts[1], parts[0]))

	// Write reminders to file.
	if err := c.WriteJSON(); err != nil {
		log.Printf("failed to write reminders JSON: %v\n", err)
	}
}

// DeleteReminder deletes an active reminder.
func DeleteReminder(s *discordgo.Session, m *discordgo.MessageCreate, c *scheduler.Scheduler, msg string) {
	id, err := strconv.Atoi(msg)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Failed to parse message to a number.")
		log.Printf("failed to parse ID to int: %v\n", err)
		return
	}

	for remind, remindID := range c.ReminderToID {
		if int(remindID) == id {
			s.ChannelMessageSend(remind.Channel, fmt.Sprintf("Removing reminder `%s`.", remind.Description))
			c.JobsMutex.Lock()
			c.Runner.Remove(remindID)
			delete(c.ReminderToID, remind)
			c.JobsMutex.Unlock()
			c.WriteJSON()
			return
		}
	}
}

// ListReminders lists active reminders.
func ListReminders(s *discordgo.Session, m *discordgo.MessageCreate, c *scheduler.Scheduler) {
	if len(c.ReminderToID) == 0 {
		s.ChannelMessageSend(m.ChannelID, "There are no active reminders.")
		return
	}

	var builder strings.Builder
	for remind, id := range c.ReminderToID {
		builder.WriteString(fmt.Sprintf("`[%d]` %s\n", id, remind.Description))
	}
	s.ChannelMessageSend(m.ChannelID, strings.TrimSpace(builder.String()))
}
