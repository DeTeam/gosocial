package main

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	qrcode "github.com/skip2/go-qrcode"
)

//go:embed templates/*
var resources embed.FS

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

type UserProfile struct {
	Handle      string
	Connections []Connection
}

func main() {
	serverOrigin := os.Getenv("SERVER_ORIGIN")

	if serverOrigin == "" {
		serverOrigin = "http://localhost:8080"
	}

	// Using in-memory sqlite
	// Tables are create on the fly
	var db Storage
	db, err := NewMemoryStorage()

	if err != nil {
		slog.Error("failed to create storage", "error", err)
		return
	}

	// Using golang html templates for a few pages
	t := &Template{
		templates: template.Must(template.ParseFS(resources, "templates/*")),
	}

	e := echo.New()
	e.Renderer = t

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index.html", nil)
	})

	// User handle is stored within "handle" cookie
	e.POST("/login", func(c echo.Context) error {
		handle := c.FormValue("handle")

		c.SetCookie(&http.Cookie{
			Name:  "handle",
			Value: handle,
		})

		// SeeOther would force POST -> GET when redirecting
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("%s/connections", serverOrigin))
	})

	// Instead of silently failing when user is not logged in we can use basic auth to get the handle.
	// This way a person following link from a qr code would be able to connect smoothly.
	loginMiddleware := middleware.BasicAuthWithConfig(middleware.BasicAuthConfig{
		Skipper: func(c echo.Context) bool {
			handle, err := c.Cookie("handle")

			if err != nil {
				return false
			}

			c.Set("handle", handle.Value)
			return true
		},

		Validator: func(username, password string, c echo.Context) (bool, error) {
			c.SetCookie(&http.Cookie{
				Name:  "handle",
				Value: username,
			})
			c.Set("handle", username)

			return true, nil
		},
	})

	e.GET("/connections", func(c echo.Context) error {
		handle, ok := c.Get("handle").(string)
		if !ok {
			return c.String(http.StatusUnauthorized, "not logged in")
		}

		connections, err := db.ListConnections(handle)

		if err != nil {
			return c.String(http.StatusInternalServerError, "failed to list connections")
		}

		profile := UserProfile{
			Handle:      handle,
			Connections: connections,
		}

		return c.Render(http.StatusOK, "connections.html", profile)
	}, loginMiddleware)

	e.POST("/invites", func(c echo.Context) error {
		handle, ok := c.Get("handle").(string)
		if !ok {
			return c.String(http.StatusUnauthorized, "not logged in")
		}

		invite, err := db.NewInvite(handle)

		if err != nil {
			return c.String(http.StatusInternalServerError, "failed to create invite")
		}

		fullURL := fmt.Sprintf("%s/qr/%s", serverOrigin, invite)

		return c.Redirect(http.StatusSeeOther, fullURL)
	}, loginMiddleware)

	e.GET("/qr/:invite", func(c echo.Context) error {
		invite := c.Param("invite")

		fullURL := fmt.Sprintf("%s/invites/%s", serverOrigin, invite)

		var png []byte
		png, err := qrcode.Encode(fullURL, qrcode.Medium, 256)

		if err != nil {
			return c.String(http.StatusInternalServerError, "failed to generate QR code")
		}

		slog.Info("QR Code Generated", "url", fullURL)

		return c.Blob(http.StatusOK, "image/png", png)
	})

	e.GET("/invites/:invite", func(c echo.Context) error {
		currentUserHandle, ok := c.Get("handle").(string)
		if !ok {
			return c.String(http.StatusUnauthorized, "not logged in")
		}

		err = db.UseInvite(currentUserHandle, c.Param("invite"))

		if err != nil {
			return c.String(http.StatusInternalServerError, "failed to use invite")
		}

		fullURL := fmt.Sprintf("%s/connections", serverOrigin)
		return c.Redirect(http.StatusTemporaryRedirect, fullURL)
	}, loginMiddleware)

	if err := e.Start(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("failed to start server", "error", err)
	}
}
