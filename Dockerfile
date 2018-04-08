# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang

# Pass in the bot token as an argument
ARG bot_token
ENV BOT_TOKEN=$bot_token

# Copy the local package files to the container's workspace.
ADD . /go/src/github.com/facecord

# Build the outyet command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
RUN go install github.com/davlia/fbmsgr
RUN go install github.com/bwmarrin/discordgo

# Run the outyet command by default when the container starts.
ENTRYPOINT /go/bin/facecord