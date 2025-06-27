# choco install make

build:
	cd ./src && go build -o ../dist/cezzis-cocktails.exe ./cmd

copyenv:
	cp ./src/.env.local ./dist/.env.local && cp ./src/.env ./dist/.env

run:
	./dist/cezzis-cocktails.exe

clean:
	rm -f ./dist/cezzis-cocktails.exe

compile: clean build copyenv

