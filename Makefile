BINARY_NAME=linkify
DIRS=shortener auth

docker:
	docker compose down && docker image prune -f && docker compose up -d --build

swag:
	cd shortener && swag init -g ./cmd/${BINARY_NAME}/main.go -o ./docs

test:
	@for dir in $(DIRS); do \
		echo "Running tests in $$dir..."; \
		cd $$dir && go test -v -race -parallel 5 -shuffle=on -coverprofile=./cover.out -covermode=atomic ./... && cd .. || exit 1; \
	done

lint:
	@for dir in $(DIRS); do \
		echo "Running linter in $$dir..."; \
		cd $$dir && golangci-lint run ./... && cd .. || exit 1; \
	done