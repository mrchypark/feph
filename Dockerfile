
FROM golang:1.14.1-buster AS builder
WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download
COPY main.go main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-w -s' -o feph main.go

FROM gcr.io/distroless/base-debian10
EXPOSE 4000
COPY --from=build /app/feph /
CMD ["/feph"]