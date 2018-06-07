package main

import (
	"encoding/json"
	"github.com/bwmarrin/discordgo"
	"github.com/robfig/cron"
	"io/ioutil"
	"log"
	"math"
	"strconv"
	"time"
)

// Constants for reminders
const (
	MaxNameLength               = 256
	ReminderLimit               = 25
	InvalidCron                 = " Cron expression is invalid. Refer to https://crontab.guru for valid expressions."
	InvalidDateParse            = " Error parsing time. Make sure it is in the format: YYYY-MM-DD."
	InvalidDescription          = " Your description is over the 256 character limit."
	NoIncomplete                = " You have no incomplete reminders in progress."
	AskDate                     = " What date is the reminder for? Respond with `!setdate <DD-MM-YYYY>`"
	TimeParseFormat             = "02-01-2006"
	DateCommandLength           = 19
	NewReminderCommandMinLength = 12
	IntervalCommandMinLength    = 14
	DescriptionCommandMinLength = 17
	DeleteCommandMinLength      = 12
	RemindersFile               = "./reminders.json"
)

// Embed : struct for Discord embeds
type Embed struct {
	*discordgo.MessageEmbed
}

// Reminder : Struct for parsing reminder input
type Reminder struct {
	Date        string    `json:"date"`
	Description string    `json:"description"`
	Interval    string    `json:"interval"`
	IsDead      bool      `json:"isDead"`
	ParsedTime  time.Time `json:"parseTime"`
}

// CreateEmbed : Creates an Embed object
func CreateEmbed() *Embed {
	return &Embed{&discordgo.MessageEmbed{}}
}

// SetTitle : Sets title of embed
func (e *Embed) SetTitle(title string) *Embed {
	e.Title = title
	return e
}

// SetDescription : Sets description of embed
func (e *Embed) SetDescription(description string) *Embed {
	e.Description = description
	return e
}

// AddField : Add field to embed
func (e *Embed) AddField(name, value string) *Embed {
	e.Fields = append(e.Fields, &discordgo.MessageEmbedField{
		Name:  name,
		Value: value,
	})
	return e
}

// SetColor : Sets color of embed
func (e *Embed) SetColor(color int) *Embed {
	e.Color = color
	return e
}

// TimeDiff : Get difference between future time and now
func TimeDiff(future time.Time) int {
	return int(math.Ceil(future.Sub(time.Now()).Hours() / 24))
}

// ValidateCron : Checks if a cron expression is valid
func ValidateCron(expression string) (cron.Schedule, bool) {
	cronParser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	schedule, err := cronParser.Parse(expression)

	if err != nil {
		return nil, false
	}

	return schedule, true
}

// AddCron : Add job to cron
func AddCron(index int, s *discordgo.Session, m *discordgo.MessageCreate) {
	// 10 seconds in cron expression to account for delays
	cronRunner.AddFunc("10 "+runningReminders[index].Interval, func() {
		days := TimeDiff(runningReminders[index].ParsedTime)
		if days > 0 && !runningReminders[index].IsDead {
			embed := CreateEmbed().
				SetColor(0x5e35b1).
				SetTitle(runningReminders[index].Description).
				SetDescription(strconv.Itoa(days) + " days remaining until " + runningReminders[index].Date + ".")

			s.ChannelMessageSendEmbed(m.ChannelID, embed.MessageEmbed)
			log.Println("Reminder \"" + runningReminders[index].Description + "\" alerted.")
		}
	})
}

// InitRemind : Alerts the new reminder and adds job to cron
func InitRemind(s *discordgo.Session, m *discordgo.MessageCreate, id string, remind Reminder) {
	embed := CreateEmbed().
		SetTitle("New reminder created").
		SetDescription("Interval: `"+remind.Interval+"`").
		AddField(remind.Description, "Date: `"+remind.Date+"`").
		SetColor(0xff00)

	s.ChannelMessageSendEmbed(m.ChannelID, embed.MessageEmbed)
	log.Println("Reminder \"" + remind.Description + "\" created with interval " +
		remind.Interval + " for " + remind.Date + ".")

	runningReminders = append(runningReminders, remind)

	WriteJSON()
	AddCron(len(runningReminders)-1, s, m)
	delete(tempStore, id)
}

// WriteJSON : Write runningReminders to file
func WriteJSON() {
	remindersJSON, _ := json.Marshal(runningReminders)

	err := ioutil.WriteFile(RemindersFile, remindersJSON, 0644)

	if err != nil {
		log.Printf("JSON write error: %v\n", err)
	}
}

// ReadJSON : Read JSON file into runningReminders
func ReadJSON() {
	file, err := ioutil.ReadFile(RemindersFile)

	if err != nil {
		log.Printf("JSON read error : %v\n", err)
	}

	json.Unmarshal(file, &runningReminders)

	for index, element := range runningReminders {
		if element.IsDead {
			runningReminders = append(runningReminders[:index], runningReminders[index+1:]...)
		}
	}

	WriteJSON()
}

// ListReminders : Create an embed of all reminders
func ListReminders() *Embed {
	embed := CreateEmbed().
		SetTitle("List of reminders").
		SetDescription("`<Reminder ID>. <Reminder description>`").
		SetColor(0xff)

	fieldDescription := ""

	for index := range runningReminders {
		if runningReminders[index].IsDead {
			fieldDescription = runningReminders[index].Description + " (Stopped)"
		} else {
			fieldDescription = runningReminders[index].Description
		}
		embed.AddField(strconv.Itoa(index+1)+". "+fieldDescription, runningReminders[index].Date+
			"("+strconv.Itoa(TimeDiff(runningReminders[index].ParsedTime))+" days remaining)")
	}

	return embed
}

// ParseInt : Parses integer from string
func ParseInt(str string) (int, error) {
	return strconv.Atoi(str)
}
