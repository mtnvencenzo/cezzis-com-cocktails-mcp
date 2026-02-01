# choco install make
# OR (on linux)
# sudo apt update
# sudo apt install build-essential

# go get -tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
# go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
# go install golang.org/x/tools/...@latest
# go install github.com/incu6us/goimports-reviser/v3@latest
# go install github.com/quantumcycle/go-ignore-cov@latest
# go install github.com/t-yuki/gocover-cobertura@latest
# cd /usr/local && sudo curl -sSfL https://raw.githubusercontent.com/dotenv-linter/dotenv-linter/master/install.sh | sudo sh -s

# https://github.com/mehdihadeli/Go-MediatR

### goimports-reviser -rm-unused -format -recursive .

# ------------------------------------------------------------
# Common Tasks
# ------------------------------------------------------------
tidy:
	cd ./cocktails.mcp/src && go mod tidy

lint:
	cd ./cocktails.mcp/src && golangci-lint run --fix && dotenv-linter fix
	
imports:
	goimports-reviser -rm-unused -format -recursive ./cocktails.mcp/src

fmt:
	cd ./cocktails.mcp/src && gofmt -s -w .

vet:
	cd ./cocktails.mcp/src && go vet ./...

test:
	cd ./cocktails.mcp/src && \
	go test \
		-cover \
		-coverprofile=../../coverage.out \
		-covermode count \
		-v ./...

cover:
	@make test
	@echo "Generating coverage reports..."
	cd ./cocktails.mcp/src && \
	go-ignore-cov --file ../../coverage.out --exclude-globs="**/test/**,cmd/**,**/*_gen.go,**/testutils/**" && \
	gocover-cobertura < ../../coverage.out > ../../cobertura.xml && \
	go tool cover -func=../../coverage.out

gen-aisearch-api:
	cd ./cocktails.mcp/src && go generate ./internal/api/aisearch

gen-cocktails-api:
	cd ./cocktails.mcp/src && go generate ./internal/api/cocktailsapi

gen-accounts-api:
	cd ./cocktails.mcp/src && go generate ./internal/api/accountsapi

standards:
	@echo "Running code standards checks..."
	@make tidy
	@make fmt
	@make imports
	@make vet
	@make lint
	@make cover
	@echo "Code standards checks completed."

update:
	cd ./cocktails.mcp/src && \
	go get -u ./...

# ------------------------------------------------------------
# build
# ------------------------------------------------------------

clean:
	rm -rf ./cocktails.mcp/dist/linux && mkdir -p ./cocktails.mcp/dist/linux

copyenv:
	@mkdir -p ./cocktails.mcp/dist/linux
	@if [ -f ./cocktails.mcp/src/.env ]; then cp ./cocktails.mcp/src/.env ./cocktails.mcp/dist/linux/.env && echo 'copied .env to dist'; fi
	@if [ -f ./cocktails.mcp/src/.env.local ]; then cp ./cocktails.mcp/src/.env.local ./cocktails.mcp/dist/linux/.env.local && echo 'copied .env.local to dist'; fi

build:
	cd ./cocktails.mcp/src && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags "-s -w" -o ../dist/linux/cezzis-cocktails ./cmd

prepare: tidy imports fmt vet lint test

compile: clean build copyenv

docker-build:
	cd ./cocktails.mcp && docker build -t cocktails-mcp -f ./Dockerfile .

compile-docker: clean build copyenv docker-build

run-docker:
	docker run --restart=always -d --name cocktails-mcp -p 3001:7999 cocktails-mcp