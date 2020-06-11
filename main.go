package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gofiber/fiber"
	"github.com/imroc/req"
)

func proxy(c *fiber.Ctx) {
	c.Accepts("application/json")
	ret := proxyOnly(c.Params("*"), c)
	var result map[string]interface{}
	json.Unmarshal(ret.Bytes(), &result)
	if len(result) == 0 {
		var result []map[string]interface{}
		json.Unmarshal(ret.Bytes(), &result)
	}
	c.Status(200).JSON(result)
}

func proxyOnly(target string, c *fiber.Ctx) *req.Resp {

	header := make(http.Header)

	c.Fasthttp.Request.Header.VisitAll(func(key, value []byte) {
		header.Set(string(key), string(value))
	})

	header.Set("X-Forwarded-Host", header.Get("Host"))

	log.Println(target)
	// fmt.Println(string(c.Fasthttp.Request.Body()))
	r, err := req.Post("http://localhost:5005/"+target, header, req.BodyJSON(string(c.Fasthttp.Request.Body())))
	if err != nil {
		log.Fatal(err)
	}
	return r
}

func main() {

	version := "chk-v0.0.6"

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

	app.Post("/*", proxy)

	app.Listen("0.0.0.0:4000")
}
