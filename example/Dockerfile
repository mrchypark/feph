FROM golang:1.14.1-buster as builder
WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-w -s' -o example main.go

FROM gcr.io/distroless/base-debian10
EXPOSE 3000
COPY --from=builder /app/example /

CMD ["./example"]