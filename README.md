## GChat bot for OpenAI ChatGPT.

- `go build -o gchatgpt.bin *.go` to build.
- Run it somewhere and register `https://your-gchatgpt-url:8000` in GSuite.

## Usage
- Build the binary: `go build -ldflags="-s -w" -o gchatgpt.bin *.go`
- Run `./gchatgpt.bin -h` to see config flags.

To run the server:
```shell
OPENAI_API_KEY="sk-xxxx" GOOGLE_PROJECT_NUMBER="123456789" ./gchatgpt.bin
```


On GChat, speak to the bot by tagging it. `@gchatgpt What is the answer to life, universe, and everything?`
