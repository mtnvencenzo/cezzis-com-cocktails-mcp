# choco install make

build:
	cd ./cocktails.mcp/src && go build -o ../dist/cezzis-cocktails.exe ./cmd

copyenv:
	cp ./cocktails.mcp/src/.env.local ./cocktails.mcp/dist/.env.local && cp ./cocktails.mcp/src/.env ./cocktails.mcp/dist/.env

run:
	./cocktails.mcp/dist/cezzis-cocktails.exe

run-http:
	./cocktails.mcp/dist/cezzis-cocktails.exe --http :8181

clean:
	rm -rf ./cocktails.mcp/dist/ && mkdir ./cocktails.mcp/dist

compile: clean build copyenv

tidy:
	cd ./cocktails.mcp/src && go mod tidy

lint:
	cd ./cocktails.mcp/src && golangci-lint run --fix

fmt:
	cd ./cocktails.mcp/src && gofmt -s -w .

