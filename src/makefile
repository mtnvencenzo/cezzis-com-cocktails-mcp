# choco install make

build:
	go build -o ../dist/cezzis-cocktails.exe ./cmd

copyenv:
	cp .env.local ../dist/.env.local && cp .env ../dist/.env

run:
	../dist/cezzis-cocktails.exe

clean:
	rm ../dist/cezzis-cocktails.exe

compile: clean build copyenv

