package web

import (
	"net/http"
	"time"

	"github.com/labstack/echo"
)

// WriteUserCookie sets user name to cookie
func WriteUserCookie(c echo.Context, name string) {
	cookie := new(http.Cookie)
	cookie.Name = "pploy_user"
	cookie.Value = name
	cookie.Expires = time.Now().Add(12 * 30 * 24 * time.Hour)
	cookie.Path = "/"
	c.SetCookie(cookie)
}

// ReadUserCookie gets user name from cookie, returns empty string when it's not set
func ReadUserCookie(c echo.Context) string {
	cookie, err := c.Cookie("pploy_user")
	if err != nil {
		return ""
	}
	return cookie.Value
}

func currentUser(c echo.Context) *string {
	u := ReadUserCookie(c)
	if u == "" {
		return nil
	}
	return &u
}
