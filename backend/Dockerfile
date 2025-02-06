FROM golang:1.20

WORKDIR /app

COPY . .

RUN go mod init pinger && go mod tidy
RUN go build -o pinger

CMD ["./pinger"]
