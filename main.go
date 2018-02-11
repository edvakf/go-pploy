package main

import (
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
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
	allProjects, err := GetAllProjects()
	if err != nil {
		return err
	}
	var project *Project
	if p := c.Param("project"); p == "" {
		project = nil
	} else {
		project = MakeCurrentProject(allProjects, p)
	}
	return c.JSON(http.StatusOK, Status{
		AllProjects:    allProjects,
		CurrentProject: project,
		AllUsers:       []string{"foo", "bar"},
		CurrentUser:    getCurrentUser(c),
	})
}

func getCurrentUser(c echo.Context) *string {
	sess, _ := session.Get("session", c)
	// sess.Values["user"] = "bar"
	// sess.Save(c.Request(), c.Response())
	if u, ok := sess.Values["user"].(string); ok {
		return &u
	}
	return nil
}

func createProject(c echo.Context) error {
	url := c.FormValue("url")
	name, err := CreateProject(url)
	if err != nil {
		return err // TODO: これどうなる？
	}
	return c.Redirect(http.StatusFound, "./"+name)
}

var SessionSecret string
var LockDuration time.Duration
var WorkDir string

func main() {
	e := echo.New()
	// e.Use(middleware.Rewrite(map[string]string{
	// 	"/*": "/assets/index.html",
	// }))
	e.Use(session.Middleware(sessions.NewCookieStore([]byte(SessionSecret))))

	e.POST("/_create", createProject)
	e.GET("/api/status/", getStatusAPI)
	e.GET("/api/status/:project", getStatusAPI)
	e.GET("/assets/*", echo.WrapHandler(http.FileServer(Assets)))
	e.GET("/*", getIndex) // rewrite middlewareでできそう
	// e.Static("/public", "/Users/atsushi/go/src/github.com/edvakf/go-pploy/public")
	e.Logger.Fatal(e.Start(":1323"))
}

func init() {
	flag.StringVar(&SessionSecret, "secret", "session-secret", "A very secret string for the cookie session store")
	flag.DurationVar(&LockDuration, "lock", 20*time.Minute, "Duration (ex. 10m) for lock gain")
	flag.StringVar(&WorkDir, "workdir", "", "Working directory")
	flag.Parse()

	if WorkDir == "" {
		panic("Please set workdir flag")
	}

	InitWorkDir(WorkDir)
}

func InitWorkDir(workDir string) {
	os.MkdirAll(workDir, os.ModePerm)
	os.MkdirAll(workDir+"/projects", os.ModePerm)
	os.MkdirAll(workDir+"/logs", os.ModePerm)
}
