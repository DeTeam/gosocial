package main

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
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
		// 2. Generate a QR code link for the invite, return in response
		return c.String(http.StatusOK, "TBD")
	})

	e.GET("/qr/:invite", func(c echo.Context) error {
		// TODO: generate a QR code png/jpeg for the given invite token
		return c.String(http.StatusOK, "TBD")
	})

	if err := e.Start(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("failed to start server", "error", err)
	}
}
