package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber"
	"github.com/gofiber/fiber/middleware"
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
	gotenv.Apply(strings.NewReader("LOG_404=true"))
}

func bodyReturn(ret *req.Resp, c *fiber.Ctx) {
	var resultList []map[string]interface{}
	if err := json.Unmarshal(ret.Bytes(), &resultList); err != nil {
		var result map[string]interface{}
		if err := json.Unmarshal(ret.Bytes(), &result); err != nil {
			c.Status(200).JSON(ret.String())
			return
		}
		c.Status(200).JSON(result)
		return
	}
	c.Status(200).JSON(resultList)
}

func proxyGet(c *fiber.Ctx) {
	target := c.Params("*")
	// request backend
	ret, err := proxyOnly(target, c)
	if err != nil {
		c.Status(404).Send("Not Found : " + "/" + target)
		return
	} else if strings.Contains(ret.String(), "Cannot") {
		c.Status(404).Send(ret.String())
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
	r, err := req.Post(turl, header, req.BodyJSON(string(c.Fasthttp.Request.Body())))
	return r, err
}

type logWriter struct {
}

func (writer logWriter) Write(bytes []byte) (int, error) {
	return fmt.Print(time.Now().UTC().Format("2006-01-02T15:04:05-0700 ") + string(bytes))
}

// "${time} ${method} ${path} - ${status} - ${latency}\nRequest :\n${body}\n",

func logs(c *fiber.Ctx) {
	log.Println(c.Method() + " " + c.Path() + "\t" + strconv.Itoa(c.Fasthttp.Response.StatusCode()))
}

func ok(state int, l404 bool, c *fiber.Ctx) {
	if state == 200 {
		c.Status(state).Send("OK")
		return
	} else {
		c.Status(state).Send("KO")
		if l404 {
			logs(c)
		}
		return
	}
}

func main() {

	version := "feph-v0.0.15"
	checkDir := os.Getenv("CHECK_DIR")
	l404, _ := strconv.ParseBool(os.Getenv("LOG_404"))

	log.SetFlags(0)
	log.SetOutput(new(logWriter))

	app := fiber.New()
	app.Use(middleware.Recover())
	app.Settings.ServerHeader = version

	// healthz
	app.Get("/", func(c *fiber.Ctx) {
		c.Status(200).Send(version)
	})
	
	app.Get("/hostname", func(c *fiber.Ctx) {
		c.Status(200).Send(os.Getenv("HOSTNAME"))
	}

	app.Get("/ext/:ext", func(c *fiber.Ctx) {
		files, err := ioutil.ReadDir(checkDir)
		if err != nil {
			ok(404, l404, c)
		}
		for _, file := range files {
			tem := strings.Split(file.Name(), ".")
			if tem[len(tem)-1] == c.Params("ext") {
				ok(200, l404, c)
			}
		}
		ok(404, l404, c)
	})

	app.Get("/filename/:name", func(c *fiber.Ctx) {
		files, err := ioutil.ReadDir(checkDir)
		if err != nil {
			ok(404, l404, c)
		}
		for _, file := range files {
			if file.Name() == c.Params("name") {
				ok(200, l404, c)
			}
		}
		ok(404, l404, c)
	})

	app.Get("/contain/:string", func(c *fiber.Ctx) {
		files, err := ioutil.ReadDir(checkDir)
		if err != nil {
			ok(404, l404, c)
		}
		for _, file := range files {
			if strings.Contains(file.Name(), c.Params("string")) {
				ok(200, l404, c)
			}
		}
		ok(404, l404, c)
	})

	app.Get("/*", proxyGet)
	app.Post("/*", proxyPost)

	app.Listen(os.Getenv("FEPH_PORT"))
}
