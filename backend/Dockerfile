FROM golang:1.21-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

RUN CGO_ENABLED=0 go build -o backend .

EXPOSE 8080

CMD ["./backend"]
