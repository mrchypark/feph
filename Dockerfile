FROM golang:1.14.1-buster AS builder
WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download
COPY main.go main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-w -s' -o rasachk main.go


FROM debian:buster-slim 
RUN apt update && apt install -y ca-certificates curl --no-install-recommends\
    && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /app/rasachk .

EXPOSE 4000
ENTRYPOINT ["/bin/sh","-c","/app/rasachk"]