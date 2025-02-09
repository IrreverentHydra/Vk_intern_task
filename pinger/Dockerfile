FROM golang:1.21-alpine

WORKDIR /app

RUN apk add --no-cache go

RUN apk add --no-cache docker-cli

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -v -o pinger || { echo "Build failed"; exit 1; }

CMD ["./pinger"]
