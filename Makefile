.PHONY: all build run clean docker-build

APP_NAME=uptime-engine
MAIN_FILE=cmd/server/main.go

all: build

build:
	CGO_ENABLED=1 go build -o $(APP_NAME) $(MAIN_FILE)

run: build
	./$(APP_NAME)

clean:
	rm -f $(APP_NAME) uptime.db

docker-build:
	docker build -t sploov/uptime-engine .

docker-run:
	docker-compose up -d
