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

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		// TODO: display a simple login page

		return c.String(http.StatusOK, "TBD")
	})

	e.GET("/user/:handle", func(c echo.Context) error {
		handle := c.Param("handle")

		// TODO: find and list connections

		return c.String(http.StatusOK, handle)
	})

	e.POST("/user/:handle/invite", func(c echo.Context) error {
		// TODO:
		// 1.	Create an invite
		// 2. Generate a QR code link for the real invite instead of mock

		invite := "test"
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
		// TODO: create a connection, invalidate the invite

		// TODO: redirect to the correct handle of the connecting user
		handle := "test"

		fullURL := fmt.Sprintf("%s/user/%s", serverOrigin, handle)
		return c.Redirect(http.StatusTemporaryRedirect, fullURL)
	})

	if err := e.Start(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("failed to start server", "error", err)
	}
}
