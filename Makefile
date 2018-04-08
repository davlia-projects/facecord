run:
	go run src/*

build:
	docker build --build-arg BOT_TOKEN=$BOT_TOKEN -t facecord --name facecord Dockerfile