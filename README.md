<a href="https://zerodha.tech"><img src="https://zerodha.tech/static/images/github-badge.svg" align="right" /></a>

## GChat bot for OpenAI ChatGPT.

`go build -o gchatgpt.bin main.go` to build or download the latest release.


## Running
- Create a new GChat app in Google Cloud Console. Copy its "Project number (app ID)"
- Build the binary `go build -ldflags="-s -w" -o gchatgpt.bin *.go` or downoad the latest release.
- Run `./gchatgpt.bin`.

Run the bot:
```shell
OPENAI_API_KEY="sk-xxxx" GOOGLE_PROJECT_NUMBER="123456789" ADDRESS=":8080" OPENAI_MODEL="" ./gchatgpt.bin
```

The bot should be listening to an https endpoing exposed to the internet. This URL should be configured in the GChat app config in the Google Cloud Console.

## Usage
On GChat, speak to the bot by tagging it. `@gchatgpt What is the answer to life, universe, and everything?`

License: MIT
