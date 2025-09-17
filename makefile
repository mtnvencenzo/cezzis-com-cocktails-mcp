# choco install make
# OR (on linux)
# sudo apt update
# sudo apt install build-essentials

# install golangci-lint
# go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

# sudo apt install golang-golang-x-tools
# sudo snap install goimports-reviser
### goimports-reviser -rm-unused -format -recursive .
# ------------------------------------------------------------
# Common Tasks
# ------------------------------------------------------------
tidy:
	cd ./cocktails.mcp/src && go mod tidy

lint:
	cd ./cocktails.mcp/src && golangci-lint run --fix
	
imports:
	goimports-reviser -rm-unused -project-name cezzis.com/cezzis-mcp-server -format -recursive ./cocktails.mcp/src

fmt:
	cd ./cocktails.mcp/src && gofmt -s -w .

test:
	cd ./cocktails.mcp/src && go test -cover -coverprofile=../../coverage.txt -v ./... &&  go tool cover -html=../../coverage.txt

# ------------------------------------------------------------
# Windows build
# ------------------------------------------------------------

clean-windows:
	rm -rf ./cocktails.mcp/dist/win && mkdir -p ./cocktails.mcp/dist/win

copyenv-windows:
	cp ./cocktails.mcp/src/.env.local ./cocktails.mcp/dist/win/.env.local && cp ./cocktails.mcp/src/.env ./cocktails.mcp/dist/win/.env

build-windows:
	cd ./cocktails.mcp/src && go build -o ../dist/win/cezzis-cocktails.exe ./cmd

run-windows:
	./cocktails.mcp/dist/win/cezzis-cocktails.exe

run-windows-http:
	./cocktails.mcp/dist/win/cezzis-cocktails.exe --http :8181

compile-windows: clean-windows build-windows copyenv-windows

# ------------------------------------------------------------
# Linux Docker build
# ------------------------------------------------------------

clean:
	rm -rf ./cocktails.mcp/dist/linux && mkdir -p ./cocktails.mcp/dist/linux

copyenv:
	cp ./cocktails.mcp/src/.env.local ./cocktails.mcp/dist/linux/.env.local && cp ./cocktails.mcp/src/.env ./cocktails.mcp/dist/linux/.env

build:
	cd ./cocktails.mcp/src && CGO_ENABLED=0 GOOS=linux go build -o ../dist/linux/cezzis-cocktails ./cmd

compile: clean build copyenv

compile-ci: clean build

docker-build:
	cd ./cocktails.mcp && docker build -t cocktails-mcp -f ./Dockerfile-CI .

compile-docker: clean build copyenv docker-build
