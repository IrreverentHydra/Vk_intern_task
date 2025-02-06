FROM golang:1.18

WORKDIR /app
COPY . .

RUN go mod init pinger && go mod tidy

CMD ["go", "run", "main.go"]
