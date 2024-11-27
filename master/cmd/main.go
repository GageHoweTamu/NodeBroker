package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3" // _ means the package is imported for its side-effects only
	"golang.org/x/crypto/bcrypt"
)

// templates for testing
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
func renderError(c echo.Context, message string) error {
	return c.HTML(400, "<p style='color: red'>"+message+"</p>")
}
func setup_db(db *sql.DB) {
	sqlStmt := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY,
        email TEXT UNIQUE,
        password_hash TEXT,
        ip_addr TEXT,
        reputation INTEGER
    );
    `
	_, err := db.Exec(sqlStmt)
	panic_on_err(err)
	fmt.Println("Created users table")
}

type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	IpAddr   string `json:"ip_addr"`
}
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type user struct {
	id         int    `json:"id"`
	email      string `json:"email"`
	ip_addr    string `json:"ip_addr"`
	reputation int    `json:"reputation"`
}
type userRequest struct {
	Email      string `json:"email"`
	IpAddr     string `json:"ip_addr"`
	Reputation int    `json:"reputation"`
}

func signup(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Handle both JSON and form data
		var email, password, ipAddr string

		// Check if this is a JSON request
		if c.Request().Header.Get("Content-Type") == "application/json" {
			req := new(SignupRequest)
			if err := c.Bind(req); err != nil {
				return c.String(400, "Invalid JSON body")
			}
			email = req.Email
			password = req.Password
			ipAddr = req.IpAddr
		} else {
			// Handle form data
			email = c.FormValue("email")
			password = c.FormValue("password")
			ipAddr = c.FormValue("ip_addr")
			if ipAddr == "" {
				ipAddr = c.RealIP() // fallback to real IP if not provided
			}
		}
		if email == "" || password == "" {
			if c.Request().Header.Get("Content-Type") == "application/json" {
				return c.String(400, "Email and password required")
			}
			return renderError(c, "Email and password required")
		}

		// Validate inputs
		if email == "" || password == "" {
			return c.String(400, "Email and password required")
		}
		if len(password) < 8 {
			return c.String(400, "Password must be at least 8 characters")
		}

		// Hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return c.String(500, "Error hashing password")
		}

		// Check if email already exists
		var exists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", email).Scan(&exists)
		if err != nil {
			return c.String(500, "Database error")
		}
		if exists {
			return c.String(400, "Email already registered")
		}

		// Insert new user
		stmt, err := db.Prepare("INSERT INTO users(email, password_hash, ip_addr, reputation) VALUES(?, ?, ?, ?)")
		if err != nil {
			return c.String(500, "Database error")
		}
		defer stmt.Close()

		result, err := stmt.Exec(email, string(hashedPassword), ipAddr, 100) // Default reputation
		if err != nil {
			return c.String(500, "Error creating user")
		}

		// Get the user's ID
		userID, err := result.LastInsertId()
		if err != nil {
			return c.String(500, "Error getting user ID")
		}

		// Create session
		sess, err := session.Get("session", c)
		if err != nil {
			return c.String(500, "Session error")
		}
		sess.Values["user_id"] = userID
		sess.Values["logged_in"] = true
		sess.Save(c.Request(), c.Response())

		// Return response based on request type
		if c.Request().Header.Get("Content-Type") == "application/json" {
			return c.JSON(200, map[string]interface{}{
				"message": "User created successfully",
				"user_id": userID,
			})
		}
		return c.Redirect(302, "/dashboard")
	}
}

func login(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Handle both JSON and form data
		var email, password string

		// Check if this is a JSON request
		if c.Request().Header.Get("Content-Type") == "application/json" {
			req := new(LoginRequest)
			if err := c.Bind(req); err != nil {
				return c.String(400, "Invalid JSON body")
			}
			email = req.Email
			password = req.Password
		} else {
			// Handle form data
			email = c.FormValue("email")
			password = c.FormValue("password")
		}

		if email == "" || password == "" {
			if c.Request().Header.Get("Content-Type") == "application/json" {
				return c.String(400, "Email and password required")
			}
			return renderError(c, "Email and password required")
		}

		// Get user from database
		var user struct {
			ID           int
			PasswordHash string
			Reputation   int
		}
		err := db.QueryRow(
			"SELECT id, password_hash, reputation FROM users WHERE email = ?",
			email,
		).Scan(&user.ID, &user.PasswordHash, &user.Reputation)
		if err != nil {
			if err == sql.ErrNoRows {
				return c.String(401, "Invalid credentials")
			}
			return c.String(500, "Database error")
		}

		// Check password
		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
		if err != nil {
			return c.String(401, "Invalid credentials")
		}

		// Create session
		sess, err := session.Get("session", c)
		if err != nil {
			return c.String(500, "Session error")
		}
		sess.Values["user_id"] = user.ID
		sess.Values["logged_in"] = true
		sess.Values["reputation"] = user.Reputation
		sess.Save(c.Request(), c.Response())

		// Return response based on request type
		if c.Request().Header.Get("Content-Type") == "application/json" {
			return c.JSON(200, map[string]interface{}{
				"message": "Login successful",
				"user_id": user.ID,
			})
		}
		return c.Redirect(302, "/dashboard")
	}
}

// Middleware to check if user is logged in
func requireLogin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, _ := session.Get("session", c)
		if sess.Values["logged_in"] != true {
			return c.String(401, "Please login first")
		}
		return next(c)
	}
}
func logout() echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get("session", c)
		if err != nil {
			return c.String(500, "Session error")
		}
		sess.Values["logged_in"] = false
		sess.Values["user_id"] = nil
		sess.Save(c.Request(), c.Response())
		return c.Redirect(302, "/login")
	}
}

// handler functions
func receive_job(c echo.Context) error {
	//handle job logic
	return c.String(200, "Job received")
}

func main() {
	server := echo.New()
	server.Use(middleware.Logger())

	// Set up templates
	server.Renderer = newTemplate()

	// Add session middleware
	store := sessions.NewCookieStore([]byte("2KQk~@sAij/ShiG"))
	server.Use(session.Middleware(store))

	db, err := sql.Open("sqlite3", "./names.db")
	setup_db(db)
	panic_on_err(err)
	defer db.Close()

	server.GET("/", func(c echo.Context) error {
		return c.String(200, "Hello, World!")
	})
	server.GET("/login", func(c echo.Context) error {
		return c.Render(200, "login.html", nil)
	})
	server.GET("/signup", func(c echo.Context) error {
		return c.Render(200, "signup.html", nil)
	})
	server.GET("/dashboard", func(c echo.Context) error {
		sess, _ := session.Get("session", c)
		if sess.Values["logged_in"] != true {
			return c.Redirect(302, "/login")
		}
		return c.Render(200, "dashboard.html", nil)
	}, requireLogin)

	// Add API routes
	server.POST("/signup", signup(db))
	server.POST("/login", login(db))
	server.POST("/logout", logout())
	server.POST("/receive-job", receive_job, requireLogin)

	server.Logger.Fatal(server.Start(":8080"))
}
