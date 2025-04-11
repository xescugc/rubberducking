.PHONY: help
help: ## Show this help
	@grep -F -h "##" $(MAKEFILE_LIST) | grep -F -v grep -F | sed -e 's/:.*##/:##/' | column -t -s '##'

.PHONY: build-game
build-game: ## Builds the internal game
	@touch assets/game
	@cd ./game && go build . && mv game ./../assets

.PHONY: run
run: build-game ## Runs the Duck!
	@go run . --verbose

.PHONY: test
test: build-game ## Tests the Duck!
	@go test ./... && cd ./game	&& go test ./...

.PHONY: build
build: build-game ## Builds the game
	@go build ./...

.PHONY: test
ctest: ## Run the tests and generate cover file
	@go test -coverprofile=cover.out ./... 
	#@cover -func=cover.out | grep total

.PHONY: test
cover: ## Run the cover tool
	@cover -html=cover.out

.PHONY: tidy
tidy: ## Runs 'go mod tidy' to bouth mains
	@go mod tidy && cd ./game && go mod tidy

.PHONY: send-message
send-message: ## Sends a message from env M
	@go run . send-message $(M)

.PHONY: build-game
install: build-game ## Installs the RubberDucking
	@go install .
