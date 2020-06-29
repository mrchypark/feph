
FROM golang:1.14.1-buster AS builder
WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download
COPY main.go main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-w -s' -o feph main.go

COPY example/main.go main.go
RUN go get github.com/gofiber/fiber
RUN go get github.com/gofiber/logger
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-w -s' -o example main.go

ENV FEPH_PORT=4000
ENV TARGET_PORT=3000
ENV CHECK_DIR=./

EXPOSE 4000

CMD ["sh","-c","./example | ./feph"]