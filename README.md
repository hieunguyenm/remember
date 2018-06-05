# remember

Discord bot for periodic reminders

## Prerequisites

It is recommended to use [`glide`](https://github.com/Masterminds/glide) to update and install dependencies but is not necessary.

Make sure to clone this repository **INSIDE** of your `$GOPATH`.

### With `glide`

```bash
glide up
glide install
```

### Without `glide`

```bash
go get github.com/bwmarrin/discordgo
go get github.com/robfig/cron
```

## Running the bot

```bash
go build
./remember <Discord bot token>
```

You will need to add the bot to your Discord server with `https://discordapp.com/oauth2/authorize?&client_id=<CLIENT_ID_HERE>&scope=bot`

Typing `!help` in your Discord chat will show you how to use the bot.
