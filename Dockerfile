
FROM golang:1.14.1-buster AS builder
WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download
COPY main.go main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-w -s' -o feph main.go

ENV FEPH_PORT=4000
ENV TARGET_PORT=5005
ENV CHECK_DIR=./

FROM gcr.io/distroless/base-debian10
EXPOSE 4000
COPY --from=build /app/feph /
CMD ["/feph"]