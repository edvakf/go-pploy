package main

import (
	"net/http"

	"github.com/labstack/echo"
)

func getIndex(c echo.Context) error {
	f, err := Assets.Open("/assets/index.html")
	if err != nil {
		return err
	}
	defer f.Close()

	return c.Stream(http.StatusOK, "text/html", f)
}

func getStatusAPI(c echo.Context) error {
	return c.JSON(http.StatusOK, Status{
		AllProjects:    []Project{},
		CurrentProject: nil,
		AllUsers:       []string{"foo", "bar"},
		CurrentUser:    nil,
	})
}

func main() {
	e := echo.New()
	// e.Use(middleware.Rewrite(map[string]string{
	// 	"/*": "/assets/index.html",
	// }))

	e.GET("/api/status/:project", getStatusAPI)
	e.GET("/assets/*", echo.WrapHandler(http.FileServer(Assets)))
	e.GET("/*", getIndex) // rewrite middlewareでできそう
	// e.Static("/public", "/Users/atsushi/go/src/github.com/edvakf/go-pploy/public")
	e.Logger.Fatal(e.Start(":1323"))
}
