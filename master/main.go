package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3" // _ means the package is imported for its side-effects only
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type user struct {
	id         string
	email      string
	ip_addr    string
	reputation string
}

// type job struct {
// 	id   int
// 	name string
// }

func receive_job(c echo.Context) error {
	//handle job logic
	return c.String(200, "Job received")
}

func main() {
	server := echo.New()
	db, err := sql.Open("sqlite3", "./names.db")
	check(err)
	fmt.Printf("db: %v", db)
	x := 0

	server.GET("/", func(c echo.Context) error {
		x++
		fmt.Printf("%dth call to / route\n", x)
		return c.String(200, "Server is running! Paths: GET /, POST /jobs")
	})

	server.POST("/receive-job", receive_job)

	server.Logger.Fatal(server.Start(":8080"))
}
