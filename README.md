# remember

Discord bot for periodic reminders

## Running the bot

### Inside the directory

```bash
go run . -token <bot token> -json=<path to new/existing JSON>
```

### Or building the binary

```bash
go build
./remember -token <bot token> -json=<path to new/existing JSON>
```

## Available commands

| Command        | Description                                                      | Syntax                                            |
| -------------- | ---------------------------------------------------------------- | ------------------------------------------------- |
| `!newreminder` | Creates a new reminder with cron interval, end date, description | `!newreminder * * * * *; 16-08-2019; description` |
| `!lsreminder`  | List active reminders with reminder ID and reminder description  | `!lsreminder`                                     |
| `!delreminder` | Delete a reminder with ID from `!lsreminder`                     | `!delreminder <reminder ID>`                      |
