package scheduler

import (
	"sync"
	"time"

	"gopkg.in/robfig/cron.v2"
)

// Scheduler contains the cron runner and the reminder -> job ID mapping.
type Scheduler struct {
	Runner       *cron.Cron
	ReminderToID map[Reminder]cron.EntryID
	JobsMutex    sync.Mutex
	JSONPath     string
}

// Reminder contains information about a reminder.
type Reminder struct {
	Date        string    `json:"date"`
	Description string    `json:"description"`
	Interval    string    `json:"interval"`
	ParsedTime  time.Time `json:"parsedTime"`
	Channel     string    `json:"channel"`
}
