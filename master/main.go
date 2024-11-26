package main

import (
	"fmt"

	"github.com/labstack/echo/v4"
)

type job struct {
    id   int
    name string
}

// Modified to match echo.HandlerFunc signature
func receive_job(c echo.Context) error {
    // Add your job handling logic here
    return c.String(200, "Job received")
}

func main() {
    server := echo.New()
    x := 0

    server.GET("/", func(c echo.Context) error {
        x++
        fmt.Printf("%dth call to / route\n", x)
        return c.String(200, "Server is running! Paths: GET /, POST /jobs")
    })

    // Move this before server.Start()
    server.POST("/receive-job", receive_job)

    server.Logger.Fatal(server.Start(":8080"))
}
