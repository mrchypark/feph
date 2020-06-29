package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gofiber/fiber"
	"github.com/gofiber/fiber/middleware"
	"github.com/gofiber/logger"
	"github.com/imroc/req"
	jsoniter "github.com/json-iterator/go"
	"github.com/subosito/gotenv"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func init() {
	// Set defualt env value
	gotenv.Apply(strings.NewReader("FEPH_PORT=4000"))
	gotenv.Apply(strings.NewReader("TARGET_PORT=5005"))
	gotenv.Apply(strings.NewReader("CHECK_DIR=./"))
}

func bodyReturn(ret *req.Resp, c *fiber.Ctx) {
	log.Println(ret.String())
	// try unmarshal json object.
	var result map[string]interface{}
	err := json.Unmarshal(ret.Bytes(), &result)
	if err != nil {
		c.Status(200).JSON(ret.String())
		return
	} else if len(result) == 0 {
		var result []map[string]interface{}
		json.Unmarshal(ret.Bytes(), &result)
		log.Println(result)
		c.Status(200).JSON(result)
		return
	}
	c.Status(200).JSON(result)
}

func proxyGet(c *fiber.Ctx) {
	target := c.Params("*")
	// request backend
	ret, err := proxyOnly(target, c)
	if strings.Contains(ret.String(), "Cannot") {
		c.Status(404).Send(ret.String())
		return
	} else if err != nil {
		c.Status(404).Send("Not Found : " + "/" + target)
		return
	}
	bodyReturn(ret, c)
}

func proxyPost(c *fiber.Ctx) {

	target := c.Params("*")
	// request backend
	ret, err := proxyWithBody(target, c)
	if strings.Contains(ret.String(), "Cannot") {
		c.Status(404).Send(ret.String())
		return
	} else if err != nil {
		c.Status(404).Send("Not Found : " + "/" + target)
		return
	}
	bodyReturn(ret, c)
}

func proxyOnly(target string, c *fiber.Ctx) (*req.Resp, error) {
	header := make(http.Header)
	c.Fasthttp.Request.Header.VisitAll(func(key, value []byte) {
		header.Set(string(key), string(value))
	})

	header.Set("X-Forwarded-Host", header.Get("Host"))
	turl := "http://localhost:" + os.Getenv("TARGET_PORT") + "/" + target
	log.Println(turl)
	r, err := req.Get(turl, header)
	return r, err
}

func proxyWithBody(target string, c *fiber.Ctx) (*req.Resp, error) {

	header := make(http.Header)
	c.Fasthttp.Request.Header.VisitAll(func(key, value []byte) {
		header.Set(string(key), string(value))
	})

	header.Set("X-Forwarded-Host", header.Get("Host"))
	turl := "http://localhost:" + os.Getenv("TARGET_PORT") + "/" + target
	log.Println(turl)
	r, err := req.Post(turl, header, req.BodyJSON(string(c.Fasthttp.Request.Body())))
	return r, err
}

func main() {

	version := "feph-v0.0.13"
	log.Println("FEPH_PORT: " + os.Getenv("FEPH_PORT"))
	log.Println("TARGET_PORT: " + os.Getenv("TARGET_PORT"))
	log.Println("CHECK_DIR: " + os.Getenv("CHECK_DIR"))

	lcfg := logger.Config{
		Format:     "${time} feph ${method} ${path} - ${status} - ${latency}\nRequest :\n${body}\n",
		TimeFormat: "2006-01-02T15:04:05-0700",
	}

	checkDir := os.Getenv("CHECK_DIR")

	app := fiber.New()
	app.Use(middleware.Recover())
	app.Use(logger.New(lcfg))
	app.Settings.ServerHeader = version

	// healthz
	app.Get("/", func(c *fiber.Ctx) {
		c.Status(200).Send(version)
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

	app.Get("/*", proxyGet)
	app.Post("/*", proxyPost)

	app.Listen(os.Getenv("FEPH_PORT"))
}
