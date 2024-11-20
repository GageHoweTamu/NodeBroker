FROM golang:1.21-alpine

RUN apk add --no-cache git

RUN git clone https://github.com/GageHoweTamu/NodeBroker

WORKDIR /NodeBroker/tcp/server

RUN go mod download
CMD ["go", "run", "server.go"]

EXPOSE 8080
