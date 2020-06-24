package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gofiber/fiber"
	"github.com/imroc/req"
	"github.com/subosito/gotenv"
)

func init() {
	// Set defualt env value
	gotenv.Apply(strings.NewReader("FEPH_PORT=4000"))
	gotenv.Apply(strings.NewReader("TARGET_PORT=5000"))
	gotenv.Apply(strings.NewReader("CHECK_DIR=."))
}

func proxy(c *fiber.Ctx) {
	// for now, just json body support
	c.Accepts("application/json")
	ret := proxyOnly(c.Params("*"), c)
	// try unmarshal json object.
	var result map[string]interface{}
	json.Unmarshal(ret.Bytes(), &result)
	if len(result) == 0 {
		// try unmarshal json array.
		var result []map[string]interface{}
		json.Unmarshal(ret.Bytes(), &result)
		c.Status(200).JSON(result)
		return
	}
	c.Status(200).JSON(result)
}

func proxyOnly(target string, c *fiber.Ctx) *req.Resp {

	header := make(http.Header)
	c.Fasthttp.Request.Header.VisitAll(func(key, value []byte) {
		header.Set(string(key), string(value))
	})

	header.Set("X-Forwarded-Host", header.Get("Host"))

	r, err := req.Post("http://localhost:"+os.Getenv("TARGET_PORT")+"/"+target,
		header, req.BodyJSON(string(c.Fasthttp.Request.Body())))
	if err != nil {
		log.Fatal(err)
	}
	return r
}

func main() {

	version := "feph-v0.0.11"
	checkDir := os.Getenv("CHECK_DIR")

	app := fiber.New()
	app.Settings.ServerHeader = version

	log.Println("Server start: " + version)

	// healthz
	app.Get("/", func(c *fiber.Ctx) {
		c.SendStatus(200)
	})

	app.Get("/ext/:ext", func(c *fiber.Ctx) {
		files, err := ioutil.ReadDir(checkDir)
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
		return
	})

	app.Get("/filename/:name", func(c *fiber.Ctx) {
		files, err := ioutil.ReadDir(checkDir)
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
		return
	})

	app.Get("/contain/:string", func(c *fiber.Ctx) {
		files, err := ioutil.ReadDir(checkDir)
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
		return
	})

	app.All("/*", proxy)

	app.Listen(os.Getenv("FEPH_PORT"))
}
