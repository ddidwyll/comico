package server

import (
	"github.com/ddidwyll/comico/cnst"
	"github.com/ddidwyll/comico/model"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"golang.org/x/crypto/acme/autocert"

	"fmt"
	"net/http"
	"strconv"
	"time"
)

var config = cnst.GetConfig()

func addCacheControl(maxAge string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Cache-Control", "public, max-age="+maxAge)
			return next(c)
		}
	}
}

func waitPlease(c echo.Context, action string, sec int64) (bool, string) {
	ua := new(model.UserAction)
	ua.Init(c.RealIP(), userId(c), action)
	return ua.IsWait(sec)
}

// USERID
func userId(c echo.Context) string {
	user, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return "without_auth"
	}
	claims := user.Claims.(jwt.MapClaims)
	id := claims["id"].(string)
	return id
}

// ACTIVITY
func activity(c echo.Context) error {
	user := userId(c)
	if user != "without_auth" {
		model.Activity(user)
	}
	return nil
}

// SUBSCRIBE
func subscribe(c echo.Context) error {
	if isWait, sec := waitPlease(c, "scribe", 1); isWait {
		return c.JSONBlob(400, []byte("{\"message\":\""+cnst.Wait(sec)+"\"}"))
	}
	user := userId(c)
	if user != "without_auth" {
		model.Tag(user, c.Param("tag"), false)
	}
	return nil
}

// IGNORE
func ignore(c echo.Context) error {
	if isWait, sec := waitPlease(c, "ignore", 1); isWait {
		return c.JSONBlob(400, []byte("{\"message\":\""+cnst.Wait(sec)+"\"}"))
	}
	user := userId(c)
	if user != "without_auth" {
		model.Tag(user, c.Param("tag"), true)
	}
	return nil
}

// MODTIMES
func modTimes(c echo.Context) error {
	t := new(model.MTime)
	t.Get()
	return c.JSON(200, t)
}

// FILES
func files(c echo.Context) error {
	str := model.GetFiles()
	if str == "" {
		return c.JSON(404, "There is no files")
	}
	return c.JSONBlob(200, []byte(str))
}

// REGISTER
func register(c echo.Context) error {
	if isWait, sec := waitPlease(c, "register", 5); isWait {
		return c.JSONBlob(400, []byte("{\"message\":\""+cnst.Wait(sec)+"\"}"))
	}
	return upsert(c, cnst.PASS, cnst.INS)
}

// LOGIN
func login(c echo.Context) error {
	if isWait, sec := waitPlease(c, "login", 2); isWait {
		return c.JSONBlob(400, []byte("{\"message\":\""+cnst.Wait(sec)+"\"}"))
	}
	p := new(model.Password)
	if err := c.Bind(p); err != nil {
		return c.JSONBlob(400,
			[]byte("{\"message\":\""+cnst.INCLOGIN+"\"}"))
	}
	u, err := p.Login()
	if err != nil {
		return c.JSONBlob(400,
			[]byte("{\"message\":\""+err.Error()+"\"}"))
	}
	token := jwt.New(jwt.SigningMethodHS256)
	exp := time.Now().Add(time.Hour * 168).Unix()
	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = u.Id
	claims["exp"] = exp
	t, err := token.SignedString([]byte(config.Secret))
	if err != nil {
		return err
	}
	return c.JSON(200, map[string]string{
		"token":  t,
		"expire": strconv.FormatInt(exp, 10),
	})
}

// UPSERT
func upsert(c echo.Context, t, action byte) error {
	if isWait, sec := waitPlease(c, fmt.Sprint("upsert", t, action), 2); isWait {
		return c.JSONBlob(400, []byte("{\"message\":\""+cnst.Wait(sec)+"\"}"))
	}
	m := model.Model(t)
	user := userId(c)
	if err := c.Bind(m); err != nil {
		return c.JSONBlob(cnst.Status(err, action+3, t))
	}
	if err := model.Upsert(t, action, m, user); err != nil {
		return c.JSONBlob(cnst.Status(err, action+3, t))
	}
	return c.JSONBlob(cnst.Status(nil, action, t))
}

// GETALL
func getAll(c echo.Context, t byte) error {
	all, err := model.GetAll(t, "")
	if err != nil {
		return c.JSON(400, err.Error())
	}
	return c.JSONBlob(200, []byte(all))
}

// DELONE
func del(c echo.Context, t byte) error {
	user := userId(c)
	err := model.Delete(t, c.Param("id"), user, false)
	if err != nil {
		return c.JSONBlob(cnst.Status(err, cnst.NTDEL, t))
	}
	return c.JSONBlob(cnst.Status(nil, cnst.DEL, t))
}

// RENEW
func renew(c echo.Context, t byte) error {
	if isWait, sec := waitPlease(c, fmt.Sprint("renew", t), 1); isWait {
		return c.JSONBlob(400, []byte("{\"message\":\""+cnst.Wait(sec)+"\"}"))
	}
	lastId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || lastId > (time.Now().Unix()-24*3600) {
		return c.JSONBlob(cnst.Status(err, cnst.DENY, t))
	}
	user := userId(c)
	err = model.Delete(t, c.Param("id"), user, true)
	if err != nil {
		return c.JSONBlob(cnst.Status(err, cnst.NTUPD, t))
	}
	return c.JSONBlob(cnst.Status(nil, cnst.UPD, t))
}

// UPLOAD
func upload(c echo.Context) error {
	if isWait, sec := waitPlease(c, "upload", 3); isWait {
		return c.JSONBlob(400, []byte("{\"message\":\""+cnst.Wait(sec)+"\"}"))
	}
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSONBlob(400, []byte("{\"message\":\"Wrong file\"}"))
	}
	if err := model.Upload(c.Param("type"), c.FormValue("name"), userId(c), file); err != nil {
		return c.JSONBlob(400, []byte("{\"message\":\""+err.Error()+"\"}"))
	}
	return c.JSONBlob(200, []byte("{\"message\":\"Image uploaded\"}"))
}

func getter(c echo.Context) error {
	switch c.Param("type") {
	case "files":
		return files(c)
	case "mtimes":
		return modTimes(c)
	case "users":
		return getAll(c, cnst.USER)
	case "goods":
		return getAll(c, cnst.GOOD)
	case "posts":
		return getAll(c, cnst.POST)
	case "cmnts":
		return getAll(c, cnst.CMNT)
	default:
		return c.JSON(404, "Not Found")
	}
}

func updater(c echo.Context) error {
	switch c.Param("type") {
	case "users":
		return upsert(c, cnst.USER, cnst.UPD)
	case "goods":
		return upsert(c, cnst.GOOD, cnst.UPD)
	case "posts":
		return upsert(c, cnst.POST, cnst.UPD)
	case "pass":
		return upsert(c, cnst.PASS, cnst.UPD)
	default:
		return c.JSON(404, "Not Found")
	}
}

func inserter(c echo.Context) error {
	switch c.Param("type") {
	case "users":
		return upsert(c, cnst.USER, cnst.INS)
	case "goods":
		return upsert(c, cnst.GOOD, cnst.INS)
	case "posts":
		return upsert(c, cnst.POST, cnst.INS)
	case "cmnts":
		return upsert(c, cnst.CMNT, cnst.INS)
	default:
		return c.JSON(404, "Not Found")
	}
}

func deleter(c echo.Context) error {
	switch c.Param("type") {
	case "users":
		return del(c, cnst.USER)
	case "goods":
		return del(c, cnst.GOOD)
	case "posts":
		return del(c, cnst.POST)
	case "cmnts":
		return del(c, cnst.CMNT)
	default:
		return c.JSON(404, "Not Found")
	}
}

func renewer(c echo.Context) error {
	switch c.Param("type") {
	case "goods":
		return renew(c, cnst.GOOD)
	case "posts":
		return renew(c, cnst.POST)
	default:
		return c.JSON(404, "Not Found")
	}
}

func Start() {
	e := echo.New()
	e.Pre(middleware.NonWWWRedirect())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			echo.GET,
			echo.PUT,
			echo.POST,
			echo.DELETE},
	}))
	s := e.Group("/", addCacheControl("3600"))
	s.Use(middleware.Static(cnst.STATIC))
	f := e.Group("/fonts", addCacheControl("31536000"))
	f.Static("/", cnst.STATIC+"fonts")
	r := e.Group("/api")
	p := e.Group("/pub")
	r.Use(middleware.JWT([]byte(config.Secret)))

	p.GET("/:type", getter)
	p.POST("/login", login)
	p.POST("/pass", register)
	r.GET("/activity", activity)
	r.GET("/subscribe/:tag", subscribe)
	r.GET("/ignore/:tag", ignore)
	r.POST("/:type", inserter)
	r.PUT("/:type", updater)
	r.DELETE("/:type/:id", deleter)
	r.GET("/renew/:type/:id", renewer)
	r.POST("/upload/:type", upload)
	if config.Port != "" && config.Port != "80" && config.Port != "443" {
		e.Logger.Fatal(e.Start(config.Host + ":" + config.Port))
	} else {
		https := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(config.Host, "www."+config.Host),
			Cache:      autocert.DirCache("./.cache"),
			ForceRSA:   true,
		}
		e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
			XSSProtection:         "1; mode=block",
			ContentTypeNosniff:    "nosniff",
			XFrameOptions:         "DENY",
			HSTSMaxAge:            31536000,
			ContentSecurityPolicy: "default-src 'self' 'unsafe-inline'; img-src * data:",
		}))
		e.Listener = https.Listener()
		go http.ListenAndServe(":80", https.HTTPHandler(nil))
		e.Logger.Fatal(e.Start(""))
	}
}
