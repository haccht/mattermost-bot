# mattermost-bot
mattermost bot framework in golang.

## Configuration

Store configuration value in envorinment variables. 
I recommend to write all them in .env file.

```
MMBOT_ENDPOINT="http://localhost:8065"
MMBOT_WEBHOOK="<your incomming webhook id>"
MMBOT_ACCOUNT="<your_mattermost_login_id@example.com>"
MMBOT_PASSWORD="<your mattermost login password>"
MMBOT_TEAMNAME="<your mattermost team>"
```

## Building an example bot

Pull this repository and build with the following command.

```
go build -o examplebot cmd/main.go
```

And run.

```
./examplebot
```

## Making a custom plugin

See sample plugins in the `plugins` directory.  
Don't forget to put your custom plugin to `cmd/main.go` before building the bot.
