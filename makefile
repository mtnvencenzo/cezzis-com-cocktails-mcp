# choco install make
# ------------------------------------------------------------
# Comman Tasks
# ------------------------------------------------------------

clean:
	rm -rf ./cocktails.mcp/dist/ && mkdir ./cocktails.mcp/dist

tidy:
	cd ./cocktails.mcp/src && go mod tidy

lint:
	cd ./cocktails.mcp/src && golangci-lint run --fix

fmt:
	cd ./cocktails.mcp/src && gofmt -s -w .

# ------------------------------------------------------------
# Windows build
# ------------------------------------------------------------

copyenv-windows:
	cp ./cocktails.mcp/src/.env.local ./cocktails.mcp/dist/win/.env.local && cp ./cocktails.mcp/src/.env ./cocktails.mcp/dist/win/.env

build-windows:
	cd ./cocktails.mcp/src && go build -o ../dist/win/cezzis-cocktails.exe ./cmd

run-windows:
	./cocktails.mcp/dist/cezzis-cocktails.exe

run-windows-http:
	./cocktails.mcp/dist/win/cezzis-cocktails.exe --http :8181

compile-windows: clean build-windows copyenv-windows

# ------------------------------------------------------------
# Linux Docker build
# ------------------------------------------------------------

copyenv-linux:
	cp ./cocktails.mcp/src/.env.local ./cocktails.mcp/dist/linux/.env.local && cp ./cocktails.mcp/src/.env ./cocktails.mcp/dist/linux/.env


build-linux:
	cd ./cocktails.mcp/src && CGO_ENABLED=0 GOOS=linux go build -o ../dist/linux/cezzis-cocktails ./cmd

docker-build:
	cd ./cocktails.mcp && docker build -t cocktails-mcp -f ./Dockerfile-CI .

compile-docker: clean build-linux copyenv-linux docker-build
