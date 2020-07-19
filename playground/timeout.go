package main

import (
	"fmt"
	"time"

	"github.com/imroc/req"
)

func main() {
	req.SetTimeout(1 * time.Second)
	res, err := req.Get("https://httpbin.org/delay/10")
	if err != nil {
		fmt.Println("err")
		fmt.Println(err)
	} else {
		fmt.Println("res")
		fmt.Println(res.String())
	}
}
