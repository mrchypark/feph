package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber"
	"github.com/gofiber/fiber/middleware"
	"github.com/imroc/req"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/subosito/gotenv"
)

func init() {
	// Set defualt env value
	gotenv.Apply(strings.NewReader("FEPH_PORT=4000"))
	gotenv.Apply(strings.NewReader("TARGET_PORT=5005"))
	gotenv.Apply(strings.NewReader("CHECK_DIR=./"))
	gotenv.Apply(strings.NewReader("LOG_LEVEL=1"))
	// set 0 means no restart.
	gotenv.Apply(strings.NewReader("TIMEOUT_RESTART=0"))
}

func main() {

	version := "feph-v0.0.19-rc2"
	checkDir := os.Getenv("CHECK_DIR")
	switch os.Getenv("LOG_LEVEL") {
	case "5":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	case "4":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "3":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "2":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "1":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "0":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "-1":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	}

	proxyTimeout, _ := strconv.Atoi(os.Getenv("TIMEOUT_RESTART"))

	if proxyTimeout >= 0 {
		setKillTimeout(proxyTimeout)
	}

	app := fiber.New()
	app.Use(middleware.Recover())
	app.Settings.ServerHeader = version
	app.Settings.StrictRouting = true
	app.Settings.CaseSensitive = true

	// healthz
	app.Get("/", func(c *fiber.Ctx) {
		c.Status(200).Send(version)
		debug(c)
	})

	app.Get("/ext/:ext", func(c *fiber.Ctx) {
		files, err := ioutil.ReadDir(checkDir)
		if err != nil {
			c.Status(404).Send("KO")
			info(c)
		} else {
			chk := false
			for _, file := range files {
				tem := strings.Split(file.Name(), ".")
				if tem[len(tem)-1] == c.Params("ext") {
					chk = true
				}
			}
			if chk {
				c.Status(200).Send("OK")
				debug(c)
			} else {
				c.Status(404).Send("KO")
				info(c)
			}
		}
	})

	app.Get("/filename/:name", func(c *fiber.Ctx) {
		files, err := ioutil.ReadDir(checkDir)
		if err != nil {
			c.Status(404).Send("KO")
			info(c)
		} else {
			chk := false
			for _, file := range files {
				if file.Name() == c.Params("name") {
					chk = true
				}
			}
			if chk {
				c.Status(200).Send("OK")
				debug(c)
			} else {
				c.Status(404).Send("KO")
				info(c)
			}
		}
	})

	app.Get("/contain/:string", func(c *fiber.Ctx) {
		files, err := ioutil.ReadDir(checkDir)
		if err != nil {
			c.Status(404).Send("KO")
			info(c)
		} else {
			chk := false
			for _, file := range files {
				if strings.Contains(file.Name(), c.Params("string")) {
					chk = true
				}
			}
			if chk {
				c.Status(200).Send("OK")
				debug(c)
			} else {
				c.Status(404).Send("KO")
				info(c)
			}
		}
	})

	app.All("/*", proxys)

	app.Listen(os.Getenv("FEPH_PORT"))

}

func info(c *fiber.Ctx) {
	log.Info().Str("path", c.Path()).
		Str("method", c.Method()).
		Str("status", strconv.Itoa(c.Fasthttp.Response.StatusCode())).
		Str("system", "feph").
		Send()
}

func debug(c *fiber.Ctx) {
	log.Debug().Str("path", c.Path()).
		Str("method", c.Method()).
		Str("status", strconv.Itoa(c.Fasthttp.Response.StatusCode())).
		Str("system", "feph").
		Send()
}

func bodyReturn(ret *req.Resp, c *fiber.Ctx) {
	var resultList []map[string]interface{}
	if err := json.Unmarshal(ret.Bytes(), &resultList); err != nil {
		var result map[string]interface{}
		if err := json.Unmarshal(ret.Bytes(), &result); err != nil {
			c.Status(200).JSON(ret.String())
			debug(c)
		} else {
			c.Status(200).JSON(result)
			debug(c)
		}
	} else {
		c.Status(200).JSON(resultList)
		debug(c)
	}
}

func proxys(c *fiber.Ctx) {
	if len(c.Body()) > 0 {
		proxyPost(c)
	} else {
		proxyGet(c)
	}
}

func proxyGet(c *fiber.Ctx) {
	target := c.Params("*")

	ret, err := proxyOnly(target, c)
	if err != nil {
		if strings.Contains(err.Error(), "context deadline exceeded") {
			kill()
		} else {
			c.Status(404).Send("Not Found : " + c.Method() + " /" + target)
			info(c)
		}
	} else if strings.Contains(ret.String(), "Cannot") {
		c.Status(404).Send(ret.String())
		info(c)
	} else {
		bodyReturn(ret, c)
	}
}

func proxyPost(c *fiber.Ctx) {
	target := c.Params("*")
	ret, err := proxyWithBody(target, c)
	if err != nil {
		if strings.Contains(err.Error(), "context deadline exceeded") {
			kill()
		} else {
			c.Status(404).Send("Not Found : " + c.Method() + " /" + target)
			info(c)
		}
	} else if strings.Contains(ret.String(), "Cannot") {
		c.Status(404).Send(ret.String())
		info(c)
	} else {
		bodyReturn(ret, c)
	}
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

func setKillTimeout(timeout int) {
	req.SetTimeout(time.Duration(timeout) * time.Second)
}

// https://pracucci.com/graceful-shutdown-of-kubernetes-pods.html
func kill() {
	log.Info().Str("timeout", os.Getenv("TIMEOUT_RESTART")).Send()
	syscall.Kill(syscall.Getpid(), syscall.SIGKILL)
}
