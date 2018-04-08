run:
	go run src/*

build:
	sudo docker build --build-arg bot_token=${BOT_TOKEN} -t facecord .
