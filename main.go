package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"syscall"
	"strconv"
	"strings"
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
	gotenv.Apply(strings.NewReader("TIMEOUT_RESTART=false"))
	gotenv.Apply(strings.NewReader("TRS_PATH=/healthz"))
	gotenv.Apply(strings.NewReader("TRS_METHOD=get"))
	gotenv.Apply(strings.NewReader("TRS_BODY_KEY="))
	gotenv.Apply(strings.NewReader("TRS_BODY_VALUE="))
	gotenv.Apply(strings.NewReader("TRS_HEADER_KEY="))
	gotenv.Apply(strings.NewReader("TRS_HEADER_VALUE="))
	gotenv.Apply(strings.NewReader("TRS_TYPE_INIT_DELAOY_SECONDS=0"))
	gotenv.Apply(strings.NewReader("TRS_TYPE_PERIOD_SECONDS=1"))
}

func main() {

	h, _ := strconv.ParseBool(os.Getenv("INNER_HEALTH"))
	if h {
		log.Info().Msg("health check run")
		go healthz()
	}

	version := "feph-v0.0.17"
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
		c.Status(404).Send("Not Found : " + c.Method() + " /" + target)
		info(c)
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
		c.Status(404).Send("Not Found : " + c.Method() + " /" + target)
		info(c)
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

func healthz() {
	t, _ := strconv.Atoi(os.Getenv("HEALTH_TYPE_INIT_DELAOY_SECONDS"))
	time.Sleep(time.Duration(t) * time.Second)
	for true {
		check()
		t, _ := strconv.Atoi(os.Getenv("HEALTH_TYPE_PERIOD_SECONDS"))
		time.Sleep(time.Duration(t) * time.Second)
	}
}

// kubectl exec POD_NAME -c CONTAINER_NAME /sbin/killall5

func check() {
	turl := "http://localhost:" + os.Getenv("TARGET_PORT") + os.Getenv("HEALTH_PATH")
	log.Info().Str("path", turl).Str("method",os.Getenv("HEALTH_METHOD")).Msg("chk")
	timeout, _ := strconv.Atoi(os.Getenv("HEALTH_TYPE_PERIOD_SECONDS"))
	req.SetTimeout(time.Duration(timeout) * time.Second)
	switch os.Getenv("HEALTH_METHOD") {
	case "get":
		tem, err := req.Get(turl)
		if err != nil {
			log.Info().Msg("kill")
			kill()
		}
		log.Info().Str("path", turl).Str("return", tem.String()).Send()
	case "post":
		b := os.Getenv("HEALTH_BODY")
		_, err := req.Post(turl, req.BodyJSON(&b))
		if err != nil {
			log.Info().Msg("kill")
			kill()
		}
	default:
	
	}
}
// https://pracucci.com/graceful-shutdown-of-kubernetes-pods.html
func kill() {
	log.Info().Str("timeout", os.Getenv("HEALTH_TYPE_PERIOD_SECONDS")).Send()
	syscall.Kill(syscall.Getpid(), syscall.SIGKILL)
}

func (h *headers) getFromEnv() {
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		matched,_:= regexp.MatchString("TRS_HEADER_KEY", pair[0])
		if matched {
			v := strings.Replace(pair[0], "TRS_HEADER_KEY", "TRS_HEADER_VALUE", 1)
			if v == "" {
				log.Warn().
					Str("env","header").
					Str("header key name", pair[0]).
					Str("TRS_HEADER_VALUE", v).
					Msg("found key but value env not exist.")
			}
			k := header{
				key: pair[1]
				value: os.Getenv(v)
			}
		}
		
	}
}

type headers []header

type header struct {
	key string
	value string
}