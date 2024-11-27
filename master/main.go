package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3" // _ means the package is imported for its side-effects only
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func newTemplate() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("templates/*.html")),
	}
}

func panic_on_err(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func setup_db(db *sql.DB) {
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY,
		email TEXT,
		ip_addr TEXT,
		reputation INTEGER
	);
	`
	_, err := db.Exec(sqlStmt)
	panic_on_err(err)
	fmt.Println("Created users table")
}

// handler functions
func receive_job(c echo.Context) error {
	//handle job logic
	return c.String(200, "Job received")
}

type user struct {
	id         int
	email      string
	ip_addr    string
	reputation int
}

type userRequest struct {
	Email      string `json:"email"`
	IpAddr     string `json:"ip_addr"`
	Reputation int    `json:"reputation"`
}

func add_user(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Parse the JSON request body into our struct
		req := new(userRequest)
		if err := c.Bind(req); err != nil {
			return c.String(400, "Invalid request body: "+err.Error())
		}

		// You might want to validate the data here
		if req.Email == "" {
			return c.String(400, "Email is required")
		}

		// Prepare the SQL statement
		stmt, err := db.Prepare("INSERT INTO users(email, ip_addr, reputation) values(?, ?, ?)")
		panic_on_err(err)
		defer stmt.Close()

		// Execute the statement with the request data
		result, err := stmt.Exec(req.Email, req.IpAddr, req.Reputation)
		if err != nil {
			return c.String(500, "Error adding user: "+err.Error())
		}

		// Get the auto-generated ID
		id, err := result.LastInsertId()
		if err != nil {
			return c.String(500, "Error getting user ID: "+err.Error())
		}

		newUser := user{
			id:         int(id),
			email:      req.Email,
			ip_addr:    req.IpAddr,
			reputation: req.Reputation,
		}

		fmt.Printf("Added user %v to database\n", newUser)
		return c.JSON(200, newUser)
	}
}

func main() {
	server := echo.New()
	server.Use(middleware.Logger())
	db, err := sql.Open("sqlite3", "./names.db")
	setup_db(db)
	panic_on_err(err)
	defer db.Close()
	fmt.Printf("db: %v", db)
	x := 0

	server.GET("/", func(c echo.Context) error {
		x++
		fmt.Printf("%dth call to / route\n", x)
		return c.String(200, "Server is running! Paths: GET /, POST /jobs")
	})

	server.GET("/new-user", add_user(db))

	server.POST("/receive-job", receive_job)

	server.Logger.Fatal(server.Start(":8080"))
}
