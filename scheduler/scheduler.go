package scheduler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"gopkg.in/robfig/cron.v2"
)

// NewScheduler creates a new cron job runner.
func NewScheduler(path string) *Scheduler {
	return &Scheduler{cron.New(), map[Reminder]cron.EntryID{}, sync.Mutex{}, path}
}

// LoadJSON reads the reminders JSON and loads reminders as cron jobs.
func (s *Scheduler) LoadJSON(ds *discordgo.Session, path string) error {
	// Read reminder JSON file.
	defer s.Runner.Start()
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil
	}

	// Unmarshal reminder JSON.
	var r []Reminder
	if err := json.Unmarshal(data, &r); err != nil {
		return fmt.Errorf("failed to unmarshal reminders: %v", err)
	}

	// Add each non-expired reminder to cron scheduler.
	for _, v := range r {
		if daysDifference(v.ParsedTime) > 0 {
			if err := s.ScheduleReminder(ds, &v); err != nil {
				return fmt.Errorf("failed to schedule reminder from JSON: %v", err)
			}
		}
	}

	log.Println("Loaded reminders JSON")
	return nil
}

// WriteJSON writes reminders to a JSON file.
func (s *Scheduler) WriteJSON() error {
	// Export reminders to a list.
	var r []Reminder
	s.JobsMutex.Lock()
	for k := range s.ReminderToID {
		r = append(r, k)
	}
	s.JobsMutex.Unlock()

	// Marshal reminders list.
	data, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("failed to marshal reminders list: %v", err)
	}

	// Write to file.
	if err := ioutil.WriteFile(s.JSONPath, data, 0755); err != nil {
		return fmt.Errorf("failed to write reminders to JSON file: %v", err)
	}
	return nil
}

// ScheduleReminder adds a reminder to the cron scheduler.
func (s *Scheduler) ScheduleReminder(ds *discordgo.Session, r *Reminder) error {
	id, err := s.Runner.AddFunc(r.Interval, func() {
		// Make local copy of reminder object.
		r := r
		days := daysDifference(r.ParsedTime)
		if days > 0 {
			// Notify reminder of remaining days is greater than 0.
			ds.ChannelMessageSendEmbed(r.Channel, &discordgo.MessageEmbed{
				Color:       0x5e35b1,
				Title:       r.Description,
				Description: fmt.Sprintf("%d days remaining until %s.", days, r.Date),
			})
			log.Printf("Alerted reminder %q for %s with interval %q\n", r.Description, r.Date, r.Interval)
		} else {
			// Remove reminder from scheduler if reminder date has passed.
			ds.ChannelMessageSend(r.Channel, fmt.Sprintf("Reminder `%s` is complete, stopping notifications.", r.Description))
			s.JobsMutex.Lock()
			s.Runner.Remove(s.ReminderToID[*r])
			delete(s.ReminderToID, *r)
			s.JobsMutex.Unlock()
			s.WriteJSON()
		}
	})
	if err != nil {
		return fmt.Errorf("failed to add job to cron: %v", err)
	}

	// Add reminder to mapping.
	s.JobsMutex.Lock()
	s.ReminderToID[*r] = id
	s.JobsMutex.Unlock()
	return nil
}

func daysDifference(t time.Time) int {
	return int(math.Ceil(t.Sub(time.Now()).Hours() / 24))
}
