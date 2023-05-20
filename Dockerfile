FROM golang:1.20

WORKDIR /go/src

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download -x

COPY . /go/src

RUN go build cmd/bot/main.go

CMD ["./main"]