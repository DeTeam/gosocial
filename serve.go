package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	qrcode "github.com/skip2/go-qrcode"
)

func main() {
	serverOrigin := "http://localhost:8080"

	// TODO: decide on which storage to use, how to integrate it better
	var db Storage = &MemoryStorage{}

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		// TODO: display a simple login page

		return c.String(http.StatusOK, "TBD")
	})

	e.GET("/user/:handle", func(c echo.Context) error {
		handle := c.Param("handle")

		connections, err := db.ListConnections(handle)

		if err != nil {
			return c.String(http.StatusInternalServerError, "failed to list connections")
		}

		return c.String(http.StatusOK, fmt.Sprint(connections))
	})

	e.POST("/user/:handle/invite", func(c echo.Context) error {
		handle := c.Param("handle")
		invite, err := db.NewInvite(handle)

		if err != nil {
			return c.String(http.StatusInternalServerError, "failed to create invite")
		}

		fullURL := fmt.Sprintf("%s/qr/%s", serverOrigin, invite)

		return c.String(http.StatusOK, fullURL)
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

	e.GET("/connect/:invite", func(c echo.Context) error {
		// TODO: find a way to get the current user handle
		currentUserHandle := "test2"

		err := db.UseInvite(currentUserHandle, c.Param("invite"))

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
