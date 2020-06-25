FROM golang:1.14.1-buster

RUN echo '                               \n\
package main                             \n\
import (                                 \n\
	"log"                                \n\
	"github.com/gofiber/fiber"           \n\
)                                        \n\
func main() {                            \n\
	// Fiber instance                    \n\
	app := fiber.New()                   \n\
	// Routes                            \n\
	app.Get("/hello", hello)             \n\
	// Start server                      \n\
	log.Fatal(app.Listen(3000))          \n\
}                                        \n\
// Handler                               \n\
func hello(c *fiber.Ctx) {               \n\
	c.Send("Hello, World 👋!")           \n\
}                                        \n'\
>> main.go

RUN go get github.com/gofiber/fiber

ARG FEPH_VER=v0.0.11
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

CMD ["sh", "-c", "go run main.go ; ./feph"]