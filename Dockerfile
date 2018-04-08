FROM golang

# Pass in the bot token as an argument
ARG bot_token
ENV BOT_TOKEN=$bot_token
ENV ENVIRONMENT="production"

# Copy the project
ADD . /go/src/github.com/facecord

# Dependencies
RUN go get github.com/davlia/fbmsgr
RUN go get github.com/bwmarrin/discordgo

# Build the project (tbh I should use `go install` but weird directory structures make it hard)
RUN go build -i -o /go/bin/facecord github.com/facecord/src

# Run it
ENTRYPOINT /go/bin/facecord