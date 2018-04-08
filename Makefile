run:
	go run src/main.go src/constants.go src/data.go src/facebook.go src/message.go src/proxy.go src/registry.go src/router.go src/session.go src/util.go

build:
	sudo docker build --build-arg bot_token=${BOT_TOKEN} -t facecord .
