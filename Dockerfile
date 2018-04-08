FROM golang

# Pass in the bot token as an argument
ARG bot_token
ENV BOT_TOKEN=$bot_token

# Copy the project
ADD . /go/src/github.com/facecord

# Dependencies
RUN go get github.com/davlia/fbmsgr
RUN go get github.com/bwmarrin/discordgo

# Run it
ENTRYPOINT /go/bin/facecord