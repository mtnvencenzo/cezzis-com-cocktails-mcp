# Docker Builds

## Cocktails Mcp
``` bash
docker build -t cocktails-mcp -f ./cocktails.mcp/Dockerfile .

docker run --restart=always -d --name cocktails-mcp -p 3001:7999 cocktails-mcp
```