# choco install make

# genapi:
# 	go tool oapi-codegen --package=main --generate types,client -o ./src/api/cocktailsapi.gen.go 'https://localhost:7176/scalar/v1/openapi.json'