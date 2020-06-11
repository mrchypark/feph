package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gofiber/fiber"
	"github.com/imroc/req"
)

func main() {

	version := "chk-v0.0.1"

	app := fiber.New()
	app.Settings.ServerHeader = version

	fmt.Println("Server start: " + version)

	// healthz
	app.Get("/", func(c *fiber.Ctx) {
		c.SendStatus(200)
	})

	app.Get("/ext/:ext", func(c *fiber.Ctx) {
		files, err := ioutil.ReadDir(".")
		if err != nil {
			c.Status(404).Send("KO")
			return
		}
		for _, file := range files {
			tem := strings.Split(file.Name(), ".")
			if tem[len(tem)-1] == c.Params("ext") {
				c.Status(200).Send("OK")
				return
			}
		}
		c.Status(404).Send("KO")
	})

	app.Get("/filename/:name", func(c *fiber.Ctx) {
		files, err := ioutil.ReadDir(".")
		if err != nil {
			c.Status(404).Send("KO")
			return
		}
		for _, file := range files {
			if file.Name() == c.Params("name") {
				c.Status(200).Send("OK")
				return
			}
		}
		c.Status(404).Send("KO")
	})

	app.Get("/contain/:string", func(c *fiber.Ctx) {
		files, err := ioutil.ReadDir(".")
		if err != nil {
			c.Status(404).Send("KO")
			return
		}
		for _, file := range files {
			if strings.Contains(file.Name(), c.Params("string")) {
				c.Status(200).Send("OK")
				return
			}
		}
		c.Status(404).Send("KO")
	})

	app.Post("/*", proxyRasa)

	app.Listen("0.0.0.0:4000")
}

func proxyRasa(c *fiber.Ctx) {
	c.Accepts("application/json")
	ret := proxyOnly(c.Params("*"), c)
	c.Status(200).JSON(ret.ToBytes)
}

func proxyOnly(target string, c *fiber.Ctx) *req.Resp {

	header := make(http.Header)

	c.Fasthttp.Request.Header.VisitAll(func(key, value []byte) {
		header.Set(string(key), string(value))
	})

	header.Set("X-Forwarded-Host", header.Get("Host"))

	// fmt.Println(string(c.Fasthttp.Request.Body()))
	r, err := req.Post("http://localhost:5005/"+target, header, req.BodyJSON(string(c.Fasthttp.Request.Body())))
	if err != nil {
		log.Fatal(err)
	}
	return r
}
