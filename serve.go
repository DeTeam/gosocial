package main

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	qrcode "github.com/skip2/go-qrcode"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	serverOrigin := "http://localhost:8080"

	// TODO: decide on which storage to use, how to integrate it better
	var db Storage = &MemoryStorage{}

	t := &Template{
		templates: template.Must(template.ParseGlob("templates/*.html")),
	}

	e := echo.New()
	e.Renderer = t

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index.html", nil)
	})

	e.POST("/login", func(c echo.Context) error {
		handle := c.FormValue("handle")

		c.SetCookie(&http.Cookie{
			Name:  "handle",
			Value: handle,
		})

		// SeeOther would force POST -> GET when redirecting
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("%s/connections", serverOrigin))
	})

	e.GET("/connections", func(c echo.Context) error {
		handleCookie, err := c.Cookie("handle")

		if err != nil {
			return c.String(http.StatusUnauthorized, "not logged in")
		}

		connections, err := db.ListConnections(handleCookie.Value)

		if err != nil {
			return c.String(http.StatusInternalServerError, "failed to list connections")
		}

		return c.Render(http.StatusOK, "connections.html", connections)
	})

	e.POST("/invites", func(c echo.Context) error {
		handleCookie, err := c.Cookie("handle")

		if err != nil {
			return c.String(http.StatusUnauthorized, "not logged in")
		}

		invite, err := db.NewInvite(handleCookie.Value)

		if err != nil {
			return c.String(http.StatusInternalServerError, "failed to create invite")
		}

		fullURL := fmt.Sprintf("%s/qr/%s", serverOrigin, invite)

		return c.Redirect(http.StatusSeeOther, fullURL)
	})

	e.GET("/qr/:invite", func(c echo.Context) error {
		invite := c.Param("invite")

		fullURL := fmt.Sprintf("%s/connect/%s", serverOrigin, invite)

		var png []byte
		png, err := qrcode.Encode(fullURL, qrcode.Medium, 256)

		if err != nil {
			return c.String(http.StatusInternalServerError, "failed to generate QR code")
		}

		return c.Blob(http.StatusOK, "image/png", png)
	})

	e.GET("/invites/:invite", func(c echo.Context) error {
		handleCookie, err := c.Cookie("handle")
		if err != nil {
			return c.String(http.StatusUnauthorized, "not logged in")
		}

		currentUserHandle := handleCookie.Value

		err = db.UseInvite(currentUserHandle, c.Param("invite"))

		if err != nil {
			return c.String(http.StatusInternalServerError, "failed to use invite")
		}

		fullURL := fmt.Sprintf("%s/user/%s", serverOrigin, currentUserHandle)
		return c.Redirect(http.StatusTemporaryRedirect, fullURL)
	})

	if err := e.Start(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("failed to start server", "error", err)
	}
}
