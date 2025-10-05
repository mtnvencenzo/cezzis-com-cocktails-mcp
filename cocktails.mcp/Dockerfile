FROM golang:1.25 AS builder
WORKDIR /app
COPY ../makefile ./makefile
COPY ../cocktails.mcp ./cocktails.mcp
# Install build dependencies
RUN apk add --no-cache go make
# Build the application
RUN cd ../ && make compile


FROM golang:1.25-alpine AS final
WORKDIR /
COPY ./dist/linux/cezzis-cocktails /cezzis-cocktails/cezzis-cocktails
# Copy environment files if they exist
COPY ./dist/linux/.env* ./cezzis-cocktails/

# Make sure the binary is executable
RUN chmod +x /cezzis-cocktails/cezzis-cocktails
# add non-root runtime user
RUN addgroup -S app && adduser -S -G app app
USER app

ENV ENV="local"

EXPOSE 7999

CMD ["/cezzis-cocktails/cezzis-cocktails", "--http", ":7999"]