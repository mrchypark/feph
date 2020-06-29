FROM golang:1.14.1-buster

ARG FEPH_VER=v0.0.13
RUN apt-get update && apt-get install -y curl \
    && curl -L https://glare.now.sh/mrchypark/feph@{$FEPH_VER}/feph-{$FEPH_VER}-linux-amd64.tar.gz -o feph.tar.gz \
    && tar -zxvf feph.tar.gz \
    && rm feph.tar.gz \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get purge -y --auto-remove

ENV FEPH_PORT=4000
ENV TARGET_PORT=3000
ENV CHECK_DIR=./

EXPOSE 4000

ADD main.go ./
RUN go get github.com/gofiber/fiber
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-w -s' -o example main.go

CMD ["sh","-c","./example | ./feph"]