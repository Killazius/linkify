BINARY_NAME=linkify

build:
	go build -o ${BINARY_NAME} cmd/${BINARY_NAME}/main.go

exec:
	./${BINARY_NAME}

clear:
	del ${BINARY_NAME}
run:
	make build && make exec && make clear

docker:
	docker compose down && docker image prune -f && docker compose up -d --build

swag:
	swag init -g .\cmd\${BINARY_NAME}\main.go -o ./docs

test:
	go test -v -race -parallel 5 -shuffle=on -coverprofile=./cover.out -covermode=atomic ./...

lint:
	golangci-lint run ./...


