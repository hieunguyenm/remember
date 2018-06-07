package main

import (
	"github.com/bwmarrin/discordgo"
	"time"
)

func addNew(s *discordgo.Session, m *discordgo.MessageCreate) {
	if len(m.Content) < NewReminderCommandMinLength {
		s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+
			" Reminder description is too short, try `!newremind <description>`")
		return
	} else if len(m.Content)+NewReminderCommandMinLength-1 > MaxNameLength {
		s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+InvalidDescription)
		return
	} else if len(runningReminders) > ReminderLimit {
		s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+
			" Maximum number of reminders reached.")
	}

	newRemind := Reminder{}
	newRemind.Description = m.Content[11:]
	tempStore[m.Author.ID] = newRemind

	s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+AskDate)
}

func setDate(s *discordgo.Session, m *discordgo.MessageCreate) {
	remind, ok := tempStore[m.Author.ID]

	if !ok {
		s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+NoIncomplete)
		return
	}

	if len(m.Content) != DateCommandLength {
		s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+InvalidDateParse)
		return
	}

	dateString := m.Content[9:]
	parsed, err := time.Parse(TimeParseFormat, dateString)

	if err != nil {
		s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+InvalidDateParse)
		return
	}

	if TimeDiff(parsed) < 0 {
		s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+
			" Date has already passed, please choose another date.")
		return
	}

	remind.ParsedTime = parsed
	remind.Date = dateString
	tempStore[m.Author.ID] = remind

	if remind.Interval == "" {
		s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+
			" At what interval would you like to be reminded at? Respond "+
			"with `!setinterval <cron expression>`\n"+
			"For help with cron expressions, refer to https://crontab.guru")
	} else {
		if TimeDiff(parsed) < 0 {
			s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+
				" Negligible difference in date and now, please choose another date.")
			remind.Date = ""
			remind.ParsedTime = time.Time{}
			tempStore[m.Author.ID] = remind
		} else {
			InitRemind(s, m, m.Author.ID, remind)
		}
	}
}

func setInterval(s *discordgo.Session, m *discordgo.MessageCreate) {
	remind, ok := tempStore[m.Author.ID]

	if !ok {
		s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+NoIncomplete)
		return
	}

	if len(m.Content) < IntervalCommandMinLength {
		s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+InvalidCron)
		return
	}

	_, valid := ValidateCron(m.Content[IntervalCommandMinLength-1:])

	if !valid {
		s.ChannelMessageSend(m.ChannelID, InvalidCron)
		return
	}

	remind.Interval = m.Content[IntervalCommandMinLength-1:]
	tempStore[m.Author.ID] = remind

	if remind.Date == "" {
		s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+AskDate)
	} else {
		if TimeDiff(remind.ParsedTime) < 0 {
			s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+
				" Negligible difference in date and now, please choose another date.")
			remind.Date = ""
			remind.ParsedTime = time.Time{}
			tempStore[m.Author.ID] = remind
		} else {
			InitRemind(s, m, m.Author.ID, remind)
		}
	}
}

func clearIncomplete(s *discordgo.Session, m *discordgo.MessageCreate) {
	delete(tempStore, m.Author.ID)
	s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+
		" Your incomplete reminder has been cleared.")
}

func setDescription(s *discordgo.Session, m *discordgo.MessageCreate) {
	if len(m.Content) < DescriptionCommandMinLength {
		s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+
			" Reminder description is too short, try `!setdescription <description>`")
		return
	} else if len(m.Content)+NewReminderCommandMinLength-1 > MaxNameLength {
		s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+InvalidDescription)
		return
	}

	remind, ok := tempStore[m.Author.ID]

	if !ok {
		s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+NoIncomplete)
		return
	}

	remind.Description = m.Content[DescriptionCommandMinLength-1:]
	tempStore[m.Author.ID] = remind

	s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+" Description set.")
}

func startReminders(s *discordgo.Session, m *discordgo.MessageCreate) {
	for index := range runningReminders {
		AddCron(index, s, m)
	}
	s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+" Reminders started.")
}

func showAll(s *discordgo.Session, m *discordgo.MessageCreate) {
	s.ChannelMessageSendEmbed(m.ChannelID, ListReminders().MessageEmbed)
}

func stopReminder(s *discordgo.Session, m *discordgo.MessageCreate) {
	if len(m.Content) < DeleteCommandMinLength {
		s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+
			" Command error, make sure the format is `!delremind <id>`")
		return
	}

	id, err := ParseInt(m.Content[DeleteCommandMinLength:])

	if err != nil {
		s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+
			" Invalid ID, please try again.")
		return
	}

	if id < 1 || id > ReminderLimit+1 {
		s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+
			" Invalid ID, please try again.")
		return
	}

	runningReminders[id-1].IsDead = true
	WriteJSON()
	s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+
		" Reminder stopped.")
}

func sendHelp(s *discordgo.Session, m *discordgo.MessageCreate) {
	embed := CreateEmbed().
		SetColor(0xff0000).
		SetTitle("Remember.go").
		SetDescription("Discord bot for periodic reminders").
		AddField("`!newremind <description>`", "Creates a new reminder with a description. "+
			"Note that user mentions in descriptions will not work, with the exception of `@everyone`").
		AddField("`!setdescription <description>`", "Sets the description for an incomplete reminder.").
		AddField("`!setdate <DD-MM-YYYY>`", "Sets the date for an incomplete reminder.").
		AddField("`!setinterval <cron expression>`", "Sets the reminder interval for an incomplete reminder.\n"+
			"For help with cron expressions, refer to https://crontab.guru").
		AddField("`!showreminders`", "Shows a list of all active reminders.").
		AddField("`!clearnew`", "Clears an incomplete reminder.").
		AddField("`!stopreminder <id>`", "Deactivates a reminder.")
	s.ChannelMessageSendEmbed(m.ChannelID, embed.MessageEmbed)
}
