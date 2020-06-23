FROM golang:alpine

ADD go.mod go.sum main.go /app/

WORKDIR /app

RUN go build main.go

CMD go run main.go
