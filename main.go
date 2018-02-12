package main

import (
	"flag"
	"net/http"
	"time"

	"github.com/edvakf/go-pploy/models/gitutil"
	"github.com/edvakf/go-pploy/models/locks"
	"github.com/edvakf/go-pploy/models/workdir"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	validator "gopkg.in/go-playground/validator.v9"
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

func getCommitsAPI(c echo.Context) error {
	project := c.Param("project")
	if !ProjectExists(project) {
		return echo.NewHTTPError(http.StatusNotFound, "project not found")
	}

	commits, err := gitutil.RecentCommits(workdir.ProjectDir(project))
	if err != nil {
		return err // TODO
	}

	return c.JSON(http.StatusOK, commits)
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
	project, err := CreateProject(url)
	if err != nil {
		return err // TODO: これどうなる？
	}
	return c.Redirect(http.StatusFound, PathPrefix+project)
}

type LockForm struct {
	User      string `form:"user" validate:"required"`
	Operation string `form:"operation" validate:"required,eq=gain|eq=release|eq=extend"`
}

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func postLock(c echo.Context) error {
	project := c.Param("project")
	if !ProjectExists(project) {
		return echo.NewHTTPError(http.StatusNotFound, "project not found")
	}

	lf := new(LockForm)
	if err := c.Bind(lf); err != nil {
		return err // TODO: 処理
	}
	if err := c.Validate(lf); err != nil {
		return err // TODO: 処理
	}

	if lf.Operation == "gain" {
		_, err := locks.Gain(project, lf.User, time.Now())
		if err != nil {
			return err
		}
	} else if lf.Operation == "release" {
		err := locks.Release(project, lf.User, time.Now())
		if err != nil {
			return err
		}
	} else if lf.Operation == "extend" {
		_, err := locks.Extend(project, lf.User, time.Now())
		if err != nil {
			return err
		}
	} else {
		panic("should not reach here")
	}

	sess, _ := session.Get("session", c)
	sess.Values["user"] = "bar"
	sess.Save(c.Request(), c.Response())

	return c.Redirect(http.StatusFound, PathPrefix+project)
}

var SessionSecret string
var PathPrefix string

func main() {
	e := echo.New()
	// e.Use(middleware.Rewrite(map[string]string{
	// 	"/*": "/assets/index.html",
	// }))
	e.Use(session.Middleware(sessions.NewCookieStore([]byte(SessionSecret))))

	e.Validator = &CustomValidator{validator: validator.New()}

	e.POST("/_create", createProject)
	e.GET("/api/status/", getStatusAPI)
	e.GET("/api/status/:project", getStatusAPI)
	e.GET("/api/commits/:project", getCommitsAPI)
	e.POST("/:project/lock", postLock)
	e.GET("/assets/*", echo.WrapHandler(http.FileServer(Assets)))
	e.GET("/:project", getIndex) // rewrite middlewareでできそう
	e.GET("/", getIndex)         // rewrite middlewareでできそう
	// e.Static("/public", "/Users/atsushi/go/src/github.com/edvakf/go-pploy/public")
	e.Logger.Fatal(e.Start(":1323"))
}

func init() {
	var lockDuration time.Duration
	var workDir string

	flag.StringVar(&SessionSecret, "secret", "session-secret", "A very secret string for the cookie session store")
	flag.DurationVar(&lockDuration, "lock", 10*time.Minute, "Duration (ex. 10m) for lock gain")
	flag.StringVar(&workDir, "workdir", "", "Working directory")
	flag.StringVar(&PathPrefix, "prefix", "/", "Path prefix of the app (eg. /pploy/), useful for proxied apps")
	flag.Parse()

	if workDir == "" {
		panic("Please set workdir flag")
	}

	locks.SetDuration(lockDuration)
	workdir.Init(workDir)
}
