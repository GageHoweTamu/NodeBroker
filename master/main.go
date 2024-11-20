package main

import (
	"fmt"

	"github.com/labstack/echo/v4"
)

// payment - https://claude.ai/chat/dd793f3d-b4c8-4c21-a5ce-33a2c5740e04
// legality - https://www.nolo.com/legal-encyclopedia/how-establish-sole-proprietorship-texas.html

type job struct {
	id   int
	name string
}

func receive_job() {}

func main() {

	server := echo.New()

	x := 0
	server.GET("/", func(c echo.Context) error {

		x++
		fmt.Printf("%dth call to / route\n", x)

		return c.String(200, "Server is running! Paths: GET /, POST /jobs")
	})
	server.Logger.Fatal(server.Start(":8080"))

	server.POST("recieve-job", receive_job)
}
