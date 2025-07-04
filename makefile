# choco install make

# ------------------------------------------------------------
# Common Tasks
# ------------------------------------------------------------
tidy:
	cd ./cocktails.mcp/src && go mod tidy

lint:
	cd ./cocktails.mcp/src && golangci-lint run --fix

fmt:
	cd ./cocktails.mcp/src && gofmt -s -w .

test:
	cd ./cocktails.mcp/src && go test -v ./...

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

clean-linux:
	rm -rf ./cocktails.mcp/dist/linux && mkdir -p ./cocktails.mcp/dist/linux

copyenv-linux:
	cp ./cocktails.mcp/src/.env.local ./cocktails.mcp/dist/linux/.env.local && cp ./cocktails.mcp/src/.env ./cocktails.mcp/dist/linux/.env

build-linux:
	cd ./cocktails.mcp/src && CGO_ENABLED=0 GOOS=linux go build -o ../dist/linux/cezzis-cocktails ./cmd

compile-linux-ci: clean-linux build-linux

docker-build:
	cd ./cocktails.mcp && docker build -t cocktails-mcp -f ./Dockerfile-CI .

compile-docker: clean-linux build-linux copyenv-linux docker-build
