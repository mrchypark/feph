
FROM golang:1.14.1-buster AS builder
WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download
COPY main.go main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-w -s' -o feph main.go

COPY example/go.mod .
RUN go mod download
COPY example/main.go main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-w -s' -o example main.go

ENV FEPH_PORT=4000
ENV TARGET_PORT=3000
ENV CHECK_DIR=./
ENV LOG_ALL_OR_ONLY_404=true

EXPOSE 4000

CMD ["sh","-c","./example | ./feph"]