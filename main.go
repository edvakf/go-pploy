package main

import (
	"bufio"
	"flag"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/edvakf/go-pploy/models/gitutil"
	"github.com/edvakf/go-pploy/models/locks"
	"github.com/edvakf/go-pploy/models/project"
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
	p, err := project.FromName(c.Param("project"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "project not found")
	}
	err = p.ReadReadme()
	if err != nil {
		return err
	}
	err = p.ReadDeployEnvs()
	if err != nil {
		return err
	}
	p.Lock = locks.Check(p.Name, time.Now())

	all, err := project.All()
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, Status{
		AllProjects:    all,
		CurrentProject: p,
		AllUsers:       []string{"foo", "bar"},
		CurrentUser:    getCurrentUser(c),
	})
}

func getCommitsAPI(c echo.Context) error {
	p, err := project.FromName(c.Param("project"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "project not found")
	}

	commits, err := gitutil.RecentCommits(workdir.ProjectDir(p.Name))
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
	p, err := project.Clone(url)
	if err != nil {
		return err // TODO: flashつけてトップにリダイレクト
	}
	return c.Redirect(http.StatusFound, PathPrefix+p.Name)
}

func getLogs(c echo.Context) error {
	p, err := project.FromName(c.Param("project"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "project not found")
	}

	r, err := p.LogReader(c.QueryParam("full") == "1")
	if err != nil {
		if os.IsNotExist(err) {
			return c.NoContent(http.StatusOK)
		}
		return err
	}
	defer r.Close()
	return c.Stream(http.StatusOK, echo.MIMETextPlainCharsetUTF8, r)
}

func postCheckout(c echo.Context) error {
	p, err := project.FromName(c.Param("project"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "project not found")
	}

	ref := c.FormValue("ref")

	r, err := p.Checkout(ref)
	if err != nil {
		return err
	}

	return transferEncodingChunked(c, r)
}

func postDeploy(c echo.Context) error {
	p, err := project.FromName(c.Param("project"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "project not found")
	}

	user := getCurrentUser(c)
	if user == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "user not provided")
	}

	env := c.FormValue("env")

	r, err := p.Deploy(env, *user)
	if err != nil {
		return err
	}

	return transferEncodingChunked(c, r)
}

func transferEncodingChunked(c echo.Context, r io.Reader) error {
	c.Response().Header().Set("Transfer-Encoding", "chunked")
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextPlainCharsetUTF8)
	c.Response().Header().Set("X-Content-Type-Options", "nosniff")
	c.Response().WriteHeader(http.StatusOK)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		c.Response().Write([]byte(scanner.Text() + "\n"))
		c.Response().Flush()
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func postLock(c echo.Context) error {
	p, err := project.FromName(c.Param("project"))
	if err != nil {
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
		_, err := locks.Gain(p.Name, lf.User, time.Now())
		if err != nil {
			return err
		}
	} else if lf.Operation == "release" {
		err := locks.Release(p.Name, lf.User, time.Now())
		if err != nil {
			return err
		}
	} else if lf.Operation == "extend" {
		_, err := locks.Extend(p.Name, lf.User, time.Now())
		if err != nil {
			return err
		}
	} else {
		panic("should not reach here")
	}

	sess, _ := session.Get("session", c)
	sess.Values["user"] = lf.User
	sess.Save(c.Request(), c.Response())

	return c.Redirect(http.StatusFound, PathPrefix+p.Name)
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
	e.GET("/:project/logs", getLogs)
	e.POST("/:project/checkout", postCheckout)
	e.GET("/:project/checkout", postCheckout)
	e.POST("/:project/deploy", postDeploy)
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
