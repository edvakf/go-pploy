package web

import (
	"net/http"
	"time"

	"github.com/labstack/echo"
)

// WriteFlashCookie sets flash message to cookie
func WriteFlashCookie(c echo.Context, message string) {
	cookie := new(http.Cookie)
	cookie.Name = "pploy_flash"
	cookie.Value = message
	cookie.Expires = time.Now().Add(1 * time.Minute)
	cookie.Path = "/"
	c.SetCookie(cookie)
}

// ReadFlashCookie gets flash message from cookie,
func ReadFlashCookie(c echo.Context) string {
	cookie, err := c.Cookie("pploy_flash")
	if err != nil {
		return ""
	}
	WriteFlashCookie(c, "")
	return cookie.Value
}
