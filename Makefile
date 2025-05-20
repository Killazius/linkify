BINARY_NAME=linkify
DIR=shortener


docker:
	docker compose down && docker image prune -f && docker compose up -d --build

swag:
	cd ${DIR} && swag init -g ./cmd/${BINARY_NAME}/main.go -o ./docs

test:
	cd ${DIR} && go test -v -race -parallel 5 -shuffle=on -coverprofile=./cover.out -covermode=atomic ./...

lint:
	cd ${DIR} && golangci-lint run ./...


