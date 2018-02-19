package web

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/edvakf/go-pploy/models/gitutil"
	"github.com/edvakf/go-pploy/models/ldapusers"
	"github.com/edvakf/go-pploy/models/locks"
	"github.com/edvakf/go-pploy/models/project"
	"github.com/edvakf/go-pploy/models/workdir"
	"github.com/fukata/golang-stats-api-handler"
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

func messageJSON(c echo.Context, message string) error {
	return c.JSON(http.StatusOK, struct { // 500?
		Message string `json:"message"`
	}{
		Message: message,
	})
}

func getStatusAPI(c echo.Context) error {
	// p is nil when project not found
	p, _ := project.Full(c.Param("project"))

	all, err := project.All()
	if err != nil {
		return messageJSON(c, err.Error())
	}

	users := ldapusers.All()
	if len(users) == 0 {
		users = []string{"foo", "bar"} // default value...
	}

	return c.JSON(http.StatusOK, struct {
		Message        string            `json:"message"`
		AllProjects    []project.Project `json:"allProjects"`
		CurrentProject *project.Project  `json:"currentProject"`
		AllUsers       []string          `json:"allUsers"`
		CurrentUser    *string           `json:"currentUser"`
	}{
		Message:        ReadFlashCookie(c),
		AllProjects:    all,
		CurrentProject: p,
		AllUsers:       users,
		CurrentUser:    currentUser(c),
	})
}

func getCommitsAPI(c echo.Context) error {
	p, err := project.FromName(c.Param("project"))
	if err != nil {
		return messageJSON(c, err.Error())
	}

	commits, err := gitutil.RecentCommits(workdir.ProjectDir(p.Name))
	if err != nil {
		return messageJSON(c, err.Error())
	}

	return c.JSON(http.StatusOK, commits)
}

func createProject(c echo.Context) error {
	form := new(struct {
		URL string `form:"url" validate:"required"`
	})
	err := validateForm(c, form)
	if err != nil {
		WriteFlashCookie(c, err.Error())
		return c.Redirect(http.StatusFound, PathPrefix)
	}

	p, err := project.Clone(form.URL)
	if err != nil {
		WriteFlashCookie(c, err.Error())
		return c.Redirect(http.StatusFound, PathPrefix)
	}

	return c.Redirect(http.StatusFound, PathPrefix+p.Name)
}

func getLogs(c echo.Context) error {
	p, err := project.FromName(c.Param("project"))
	if err != nil {
		return c.String(http.StatusOK, err.Error())
	}

	r, err := p.LogReader(c.QueryParam("full") == "1")
	if err != nil {
		return err
	}
	defer r.Close()
	return c.Stream(http.StatusOK, echo.MIMETextPlainCharsetUTF8, r)
}

func postCheckout(c echo.Context) error {
	p, err := project.FromName(c.Param("project"))
	if err != nil {
		return c.String(http.StatusOK, err.Error())
	}

	form := new(struct {
		Ref string `form:"ref" validate:"required"`
	})
	err = validateForm(c, form)
	if err != nil {
		return c.String(http.StatusOK, err.Error())
	}

	r, err := p.Checkout(form.Ref)
	if err != nil {
		return c.String(http.StatusOK, err.Error())
	}

	return transferEncodingChunked(c, r)
}

func postDeploy(c echo.Context) error {
	p, err := project.FromName(c.Param("project"))
	if err != nil {
		return c.String(http.StatusOK, err.Error())
	}

	user := currentUser(c)
	if user == nil {
		return c.String(http.StatusOK, "user cookie not set")
	}

	form := new(struct {
		Target string `form:"target" validate:"required"`
	})
	err = validateForm(c, form)
	if err != nil {
		return c.String(http.StatusOK, err.Error())
	}

	r, err := p.Deploy(form.Target, *user)
	if err != nil {
		return c.String(http.StatusOK, err.Error())
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

func postLock(c echo.Context) error {
	p, err := project.FromName(c.Param("project"))
	if err != nil {
		WriteFlashCookie(c, err.Error())
		return c.Redirect(http.StatusFound, PathPrefix)
	}

	form := new(struct {
		User      string `form:"user" validate:"required"`
		Operation string `form:"operation" validate:"required,eq=gain|eq=release|eq=extend"`
	})
	err = validateForm(c, form)
	if err != nil {
		WriteFlashCookie(c, err.Error())
		return c.Redirect(http.StatusFound, PathPrefix+p.Name)
	}

	if form.Operation == "gain" {
		_, err := locks.Gain(p.Name, form.User, time.Now())
		if err != nil {
			WriteFlashCookie(c, err.Error())
			return c.Redirect(http.StatusFound, PathPrefix+p.Name)
		}
	} else if form.Operation == "release" {
		err := locks.Release(p.Name, form.User, time.Now())
		if err != nil {
			WriteFlashCookie(c, err.Error())
			return c.Redirect(http.StatusFound, PathPrefix+p.Name)
		}
	} else if form.Operation == "extend" {
		_, err := locks.Extend(p.Name, form.User, time.Now())
		if err != nil {
			WriteFlashCookie(c, err.Error())
			return c.Redirect(http.StatusFound, PathPrefix+p.Name)
		}
	} else {
		panic("should not reach here")
	}

	WriteUserCookie(c, form.User)

	return c.Redirect(http.StatusFound, PathPrefix+p.Name)
}

func validateForm(c echo.Context, form interface{}) error {
	if err := c.Bind(form); err != nil {
		return err
	}
	if err := c.Validate(form); err != nil {
		return err
	}
	return nil
}

var PathPrefix string
var Port int

func Server() {
	e := echo.New()

	e.Validator = &Validator

	e.POST("/_create", createProject)
	e.GET("/api/status/", getStatusAPI)
	e.GET("/api/status/:project", getStatusAPI)
	e.GET("/api/commits/:project", getCommitsAPI)
	e.POST("/:project/lock", postLock)
	e.GET("/:project/logs", getLogs)
	e.POST("/:project/checkout", postCheckout)
	e.POST("/:project/deploy", postDeploy)
	e.GET("/assets/*", echo.WrapHandler(http.FileServer(Assets)))
	e.GET("/api/_stats", echo.WrapHandler(http.HandlerFunc(stats_api.Handler)))
	e.GET("/:project", getIndex) // rewrite middlewareでできそう
	e.GET("/", getIndex)         // rewrite middlewareでできそう
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", Port)))
}
