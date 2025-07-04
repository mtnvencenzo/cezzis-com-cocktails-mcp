# Docker Builds

## Cocktails Mcp
``` bash
docker build -t cocktails-mcp -f ./cocktails.mcp/Dockerfile-CI .

docker run --restart=always -d --name cocktails-mcp -p 3001:8181 cocktails-mcp
```