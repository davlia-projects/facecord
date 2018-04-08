run:
	go run src/*

build:
	docker build --build-arg bot_token=${BOT_TOKEN} -t dliao/facecord:latest .
	docker push dliao/facecord:latest
