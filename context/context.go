// Copyright 2019 GoAdmin Core Team. All rights reserved.
// Use of this source code is governed by a Apache-2.0 style
// license that can be found in the LICENSE file.

package context

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/GoAdminGroup/go-admin/modules/constant"
)

const abortIndex int8 = math.MaxInt8 / 2

// gin.Context(gin-gonic套件)的簡化版本
// Context is the simplify version of web framework context.
// But it is important which will be used in plugins to custom
// the request and response. And adapter will help to transform
// the Context to the web framework`s context. It has three attributes.
// Request and response are belongs to net/http package. UserValue
// is the custom key-value store of context.
type Context struct {
	Request   *http.Request
	Response  *http.Response
	UserValue map[string]interface{}
	index     int8
	handlers  Handlers
}

// 用於匹配的request and response
// Path is used in the matching of request and response. Url stores the
// raw register url. RegUrl contains the wildcard which on behalf of
// the route params.
type Path struct {
	URL    string
	Method string
}

type RouterMap map[string]Router

// 藉由參數name取得Router(struct)，Router裡有Methods([]string)及Pattern(string)
func (r RouterMap) Get(name string) Router {
	return r[name]
}

// Router包含方法模式
type Router struct {
	Methods []string
	Patten  string //url
}

// 取得Method
func (r Router) Method() string {
	return r.Methods[0]
}

// 處理URL後回傳(處理url中有:__的字串)
func (r Router) GetURL(value ...string) string {
	u := r.Patten

	// 未處理前ex:/admin/info/:__prefix/edit
	for i := 0; i < len(value); i += 2 {
		// 處理url
		u = strings.Replace(u, ":__"+value[i], value[i+1], -1)
	}
	// 處理後 ex:/admin/info/roles/edit
	return u
}

type NodeProcessor func(...Node)

type Node struct {
	Path     string
	Method   string
	Handlers []Handler
	Value    map[string]interface{}
}

// 設定Context.UserValue藉由參數key、value
// SetUserValue set the value of user context.
func (ctx *Context) SetUserValue(key string, value interface{}) {
	ctx.UserValue[key] = value
}

// 回傳Request url path
// Path return the url path.
func (ctx *Context) Path() string {
	return ctx.Request.URL.Path
}

// Abort abort the context.
// Context.index = 63
func (ctx *Context) Abort() {
	// abortIndex = 63
	ctx.index = abortIndex
}

// 執行迴圈Context.handlers[ctx.index](ctx)
// Next should be used only inside middleware.
func (ctx *Context) Next() {
	ctx.index++

	// Context.Handlers類別為[]Handler，Handler類別為func(ctx *Context)
	for s := int8(len(ctx.handlers)); ctx.index < s; ctx.index++ {
		// 執行func(ctx *Context)
		ctx.handlers[ctx.index](ctx)
	}
}

// 設置handlers(參數)至Context.Handlers
// SetHandlers set the handlers of Context.
func (ctx *Context) SetHandlers(handlers Handlers) *Context {
	ctx.handlers = handlers
	return ctx
}

// 回傳方法
// Method return the request method.
func (ctx *Context) Method() string {
	return ctx.Request.Method
}

// 設置新Context(struct)，帶有Request(請求)以及UserValue、Response(預設的slice)
// NewContext used in adapter which return a Context with request
// and slice of UserValue and a default Response.
func NewContext(req *http.Request) *Context {

	return &Context{
		Request:   req,
		UserValue: make(map[string]interface{}),
		Response: &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
		},
		index: -1,
	}
}

const (
	HeaderContentType = "Content-Type"

	HeaderLastModified    = "Last-Modified"
	HeaderIfModifiedSince = "If-Modified-Since"
	HeaderCacheControl    = "Cache-Control"
	HeaderETag            = "ETag"

	HeaderContentDisposition = "Content-Disposition"
	HeaderContentLength      = "Content-Length"
	HeaderContentEncoding    = "Content-Encoding"

	GzipHeaderValue      = "gzip"
	HeaderAcceptEncoding = "Accept-Encoding"
	HeaderVary           = "Vary"
)

// 處理Context.Request.Body並將json轉成struct
func (ctx *Context) BindJSON(data interface{}) error {
	if ctx.Request.Body != nil {
		b, err := ioutil.ReadAll(ctx.Request.Body)
		if err == nil {
			return json.Unmarshal(b, data)
		}
		return err
	}
	return errors.New("empty request body")
}

// 處理Context.Request.Body並將json轉成struct
func (ctx *Context) MustBindJSON(data interface{}) {
	if ctx.Request.Body != nil {
		b, err := ioutil.ReadAll(ctx.Request.Body)
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(b, data)
		if err != nil {
			panic(err)
		}
	}
	panic("empty request body")
}

// Write save the given status code, headers and body string into the response.
// 將狀態碼，標頭(header)及body寫入Context.Response
func (ctx *Context) Write(code int, header map[string]string, Body string) {
	ctx.Response.StatusCode = code
	for key, head := range header {
		// 加入header
		ctx.AddHeader(key, head)
	}
	ctx.Response.Body = ioutil.NopCloser(strings.NewReader(Body))
}

// JSON serializes the given struct as JSON into the response body.
// It also sets the Content-Type as "application/json".
// 轉換成JSON存至Context.Response.body
func (ctx *Context) JSON(code int, Body map[string]interface{}) {
	ctx.Response.StatusCode = code
	//設定SetContentType
	ctx.SetContentType("application/json")
	// Marshal將struct轉成json
	BodyStr, err := json.Marshal(Body)
	if err != nil {
		panic(err)
	}
	ctx.Response.Body = ioutil.NopCloser(bytes.NewReader(BodyStr))
}

// DataWithHeaders save the given status code, headers and body data into the response.
// 將code, headers and body(參數)存在Context.Response中
func (ctx *Context) DataWithHeaders(code int, header map[string]string, data []byte) {
	ctx.Response.StatusCode = code
	for key, head := range header {
		//添加標頭
		ctx.AddHeader(key, head)
	}
	ctx.Response.Body = ioutil.NopCloser(bytes.NewBuffer(data))
}

// Data writes some data into the body stream and updates the HTTP code.
// 寫入資料並更新HTTP狀態碼(Context.Response.StatusCode、Context.Response.Body)
func (ctx *Context) Data(code int, contentType string, data []byte) {
	ctx.Response.StatusCode = code
	ctx.SetContentType(contentType)
	ctx.Response.Body = ioutil.NopCloser(bytes.NewBuffer(data))
}

// Redirect add redirect url to header.
// 添加重新導向的url(參數path)至header
func (ctx *Context) Redirect(path string) {
	// http.StatusFound = 302
	ctx.Response.StatusCode = http.StatusFound
	ctx.SetContentType("text/html; charset=utf-8")
	// 添加header
	ctx.AddHeader("Location", path)
}

// HTML output html response.
// 輸出HTML，參數body保存至Context.response.Body
func (ctx *Context) HTML(code int, body string) {
	ctx.SetContentType("text/html; charset=utf-8")
	ctx.SetStatusCode(code)
	// 將參數body保存至Context.response.Body中
	ctx.WriteString(body)
}

// HTMLByte output html response.
// 輸出HTML，將body參數設置至Context.response.Body
func (ctx *Context) HTMLByte(code int, body []byte) {
	ctx.SetContentType("text/html; charset=utf-8")
	ctx.SetStatusCode(code)
	ctx.Response.Body = ioutil.NopCloser(bytes.NewBuffer(body))
}

// WriteString save the given body string into the response.
// 將參數body保存至Context.response.Body中
func (ctx *Context) WriteString(body string) {
	ctx.Response.Body = ioutil.NopCloser(strings.NewReader(body))
}

// SetStatusCode save the given status code into the response.
// 將參數設置至Context.Response.StatusCode
func (ctx *Context) SetStatusCode(code int) {
	ctx.Response.StatusCode = code
}

// SetContentType save the given content type header into the response header.
// HeaderContentType = Content-Type
// 將參數添加至Content-Type
func (ctx *Context) SetContentType(contentType string) {
	ctx.AddHeader(HeaderContentType, contentType)
}

// 設定上次修改時間
func (ctx *Context) SetLastModified(modtime time.Time) {
	if !IsZeroTime(modtime) {
		// HeaderLastModified = Last-Modified
		// 添加header
		ctx.AddHeader(HeaderLastModified, modtime.UTC().Format(http.TimeFormat)) // or modtime.UTC()?
	}
}

var unixEpochTime = time.Unix(0, 0)

// IsZeroTime reports whether t is obviously unspecified (either zero or Unix()=0).
// 是否time is 0
func IsZeroTime(t time.Time) bool {
	return t.IsZero() || t.Equal(unixEpochTime)
}

// ParseTime parses a time header (such as the Date: header),
// trying each forth formats
// that are allowed by HTTP/1.1:
// time.RFC850, and time.ANSIC.
// 解析時間
var ParseTime = func(text string) (t time.Time, err error) {
	t, err = time.Parse(http.TimeFormat, text)
	if err != nil {
		return http.ParseTime(text)
	}

	return
}

// 清除Context.Response.Header裡的Content-Type、Content-Length、Last-Modified
func (ctx *Context) WriteNotModified() {
	// RFC 7232 section 4.1:
	// a sender SHOULD NOT generate representation metadata other than the
	// above listed fields unless said metadata exists for the purpose of
	// guiding cache updates (e.g.," Last-Modified" might be useful if the
	// response does not have an ETag field).
	delete(ctx.Response.Header, HeaderContentType)
	delete(ctx.Response.Header, HeaderContentLength)
	if ctx.Headers(HeaderETag) != "" {
		delete(ctx.Response.Header, HeaderLastModified)
	}
	// 304(未修改)
	ctx.SetStatusCode(http.StatusNotModified)
}

// 檢查是否修改過
func (ctx *Context) CheckIfModifiedSince(modtime time.Time) (bool, error) {
	if method := ctx.Method(); method != http.MethodGet && method != http.MethodHead {
		return false, errors.New("skip: method")
	}
	//HeaderIfModifiedSince = If-Modified-Since
	ims := ctx.Headers(HeaderIfModifiedSince)
	if ims == "" || IsZeroTime(modtime) {
		return false, errors.New("skip: zero time")
	}
	t, err := ParseTime(ims)
	if err != nil {
		return false, errors.New("skip: " + err.Error())
	}
	// sub-second precision, so
	// use mtime < t+1s instead of mtime <= t to check for unmodified.
	if modtime.UTC().Before(t.Add(1 * time.Second)) {
		return false, nil
	}
	return true, nil
}

// LocalIP return the request client ip.
// 回傳用戶端IP
func (ctx *Context) LocalIP() string {
	xForwardedFor := ctx.Request.Header.Get("X-Forwarded-For")
	ip := strings.TrimSpace(strings.Split(xForwardedFor, ",")[0])
	if ip != "" {
		return ip
	}

	ip = strings.TrimSpace(ctx.Request.Header.Get("X-Real-Ip"))
	if ip != "" {
		return ip
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(ctx.Request.RemoteAddr)); err == nil {
		return ip
	}

	return "127.0.0.1"
}

// SetCookie save the given cookie obj into the response Set-Cookie header.
// 設置cookie在response header Set-Cookie中
func (ctx *Context) SetCookie(cookie *http.Cookie) {
	if v := cookie.String(); v != "" {
		ctx.AddHeader("Set-Cookie", v)
	}
}

// Query get the query parameter of url.
// 取得Request url(在url中)裡的參數(key)
func (ctx *Context) Query(key string) string {
	return ctx.Request.URL.Query().Get(key)
}

// QueryDefault get the query parameter of url. If it is empty, return the default.
// 取得url的查詢參數(類似過濾條件)，如果為空返回預設值(第二個參數參數)
func (ctx *Context) QueryDefault(key, def string) string {
	value := ctx.Query(key)
	if value == "" {
		return def
	}
	return value
}

// Headers get the value of request headers key.
// 藉由參數key獲得Header
func (ctx *Context) Headers(key string) string {
	return ctx.Request.Header.Get(key)
}

// FormValue get the value of request form key.
// 藉由參數key取得multipart/form-data中的值
func (ctx *Context) FormValue(key string) string {
	return ctx.Request.FormValue(key)
}

// PostForm get the values of request form.
// 取得表單的值(所有)，將參數放於multipart/form-data.
func (ctx *Context) PostForm() url.Values {
	_ = ctx.Request.ParseMultipartForm(32 << 20)
	return ctx.Request.PostForm
}

// 判斷method是否為get以及header裡包含accept:html
func (ctx *Context) WantHTML() bool {
	return ctx.Method() == "GET" && strings.Contains(ctx.Headers("Accept"), "html")
}

// 判斷header裡包含accept:json
func (ctx *Context) WantJSON() bool {
	return strings.Contains(ctx.Headers("Accept"), "json")
}

// AddHeader adds the key, value pair to the header.
// 將參數(key、value)添加header中(Context.Response.Header)
func (ctx *Context) AddHeader(key, value string) {
	ctx.Response.Header.Add(key, value)
}

// PjaxUrl add pjax url header.
// 添加pjax url header
func (ctx *Context) PjaxUrl(url string) {
	ctx.Response.Header.Add(constant.PjaxUrlHeader, url)
}

// IsPjax check request is pjax or not.
// 判斷是否header X-PJAX:true
func (ctx *Context) IsPjax() bool {
	// constant.PjaxHeader = X-PJAX
	// 藉由參數key獲得鍵的value
	return ctx.Headers(constant.PjaxHeader) == "true"
}

// SetHeader set the key, value pair to the header.
// 設定header
func (ctx *Context) SetHeader(key, value string) {
	ctx.Response.Header.Set(key, value)
}

// 取得Context.Request.Header
func (ctx *Context) GetContentType() string {
	return ctx.Request.Header.Get("")
}

// User return the current login user.
// 回傳目前登入的用戶(Context.UserValue["user"])
func (ctx *Context) User() interface{} {
	return ctx.UserValue["user"]
}

// ServeContent serves content, headers are autoset
// receives three parameters, it's low-level function, instead you can use .ServeFile(string,bool)/SendFile(string,string)
//
// You can define your own "Content-Type" header also, after this function call
// Doesn't implements resuming (by range), use ctx.SendFile instead
// 自動設置headers，可以自定義Content-Type，讀取內容並設置至Context.Response.Body 
func (ctx *Context) ServeContent(content io.ReadSeeker, filename string, modtime time.Time, gzipCompression bool) error {
	// 檢查是否修改過
	if modified, err := ctx.CheckIfModifiedSince(modtime); !modified && err == nil {
		// 清除Context.Response.Header裡的Content-Type、Content-Length、Last-Modified
		ctx.WriteNotModified()
		return nil
	}

	// 取得Context.Request.Header
	if ctx.GetContentType() == "" {
		// 將參數添加至Content-Type
		ctx.SetContentType(filename)
	}

	buf, _ := ioutil.ReadAll(content)
	ctx.Response.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
	return nil
}

// ServeFile serves a view file, to send a file ( zip for example) to the client you should use the SendFile(serverfilename,clientfilename)
func (ctx *Context) ServeFile(filename string, gzipCompression bool) error {
	// open file
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("%d", http.StatusNotFound)
	}
	defer func() {
		_ = f.Close()
	}()
	// 回傳檔案的描述結構
	fi, _ := f.Stat()
	// 檢查是否有檔案
	if fi.IsDir() {
		return ctx.ServeFile(path.Join(filename, "index.html"), gzipCompression)
	}

	return ctx.ServeContent(f, fi.Name(), fi.ModTime(), gzipCompression)
}

type HandlerMap map[Path]Handlers

// App is the key struct of the package. App as a member of plugin
// entity contains the request and the corresponding handler. Prefix
// is the url prefix and MiddlewareList is for control flow.
type App struct {
	Requests    []Path //Path包含url、method
	Handlers    HandlerMap // HandlerMap是 map[Path]Handlers，Handlers類別為[]Handler，Handler類別為func(ctx *Context)
	Middlewares Handlers
	Prefix      string

	Routers    RouterMap // RouterMap類別為map[string]Router，Router(struct)裡有methods、patten
	routeIndex int
	routeANY   bool
}

// NewApp return an empty app.
// 回傳新的app(struct)，空的
func NewApp() *App {
	return &App{
		Requests:    make([]Path, 0),
		Handlers:    make(HandlerMap),
		Prefix:      "/",
		Middlewares: make([]Handler, 0),
		routeIndex:  -1,
		Routers:     make(RouterMap),
	}
}

// Handler defines the handler used by the middleware as return value.
type Handler func(ctx *Context) // 用於middleware 

// Handlers is the array of Handler
type Handlers []Handler

// AppendReqAndResp stores the request info and handle into app.
// support the route parameter. The route parameter will be recognized as
// wildcard store into the RegUrl of Path struct. For example:
//
//         /user/:id      => /user/(.*)
//         /user/:id/info => /user/(.*?)/info
//
// The RegUrl will be used to recognize the incoming path and find
// the handler.
// 在App(struct)新增Requests([]Path)路徑、Middlewares至Handlers
func (app *App) AppendReqAndResp(url, method string, handler []Handler) {

	//新增Path
	app.Requests = append(app.Requests, Path{
		URL:    join(app.Prefix, url),
		Method: method,
	})
	app.routeIndex++

	// 新增Middlewares至Handlers
	app.Handlers[Path{
		URL:    join(app.Prefix, url),
		Method: method,
	}] = append(app.Middlewares, handler...)
}

// Find is public helper method for findPath of tree.
// 尋找app.Handlers(map[Path]Handlers)中匹配的Path(struct)
func (app *App) Find(url, method string) []Handler {
	app.routeANY = false
	return app.Handlers[Path{URL: url, Method: method}]
}

// POST is a shortcut for app.AppendReqAndResp(url, "post", handler).
// POST等於在AppendReqAndResp(url, "post", handler)，在App(struct)新增Requests([]Path)路徑、Middlewares至Handlers
func (app *App) POST(url string, handler ...Handler) *App {
	app.routeANY = false
	app.AppendReqAndResp(url, "post", handler)
	return app
}

// GET is a shortcut for app.AppendReqAndResp(url, "get", handler).
// GET等於在AppendReqAndResp(url, "get", handler)，在App(struct)新增Requests([]Path)路徑、Middlewares至Handlers
func (app *App) GET(url string, handler ...Handler) *App {
	app.routeANY = false
	app.AppendReqAndResp(url, "get", handler)
	return app
}

// DELETE is a shortcut for app.AppendReqAndResp(url, "delete", handler).
// DELETE等於在AppendReqAndResp(url, "delete", handler)，在App(struct)新增Requests([]Path)路徑、Middlewares至Handlers
func (app *App) DELETE(url string, handler ...Handler) *App {
	app.routeANY = false
	app.AppendReqAndResp(url, "delete", handler)
	return app
}

// PUT is a shortcut for app.AppendReqAndResp(url, "put", handler).
// PUT等於在AppendReqAndResp(url, "put", handler)，在App(struct)新增Requests([]Path)路徑、Middlewares至Handlers
func (app *App) PUT(url string, handler ...Handler) *App {
	app.routeANY = false
	app.AppendReqAndResp(url, "put", handler)
	return app
}

// OPTIONS is a shortcut for app.AppendReqAndResp(url, "options", handler).
// OPTIONS等於在AppendReqAndResp(url, "options", handler)，在App(struct)新增Requests([]Path)路徑、Middlewares至Handlers
func (app *App) OPTIONS(url string, handler ...Handler) *App {
	app.routeANY = false
	app.AppendReqAndResp(url, "options", handler)
	return app
}

// HEAD is a shortcut for app.AppendReqAndResp(url, "head", handler).
// HEAD等於在AppendReqAndResp(url, "head", handler)，在App(struct)新增Requests([]Path)路徑、Middlewares至Handlers
func (app *App) HEAD(url string, handler ...Handler) *App {
	app.routeANY = false
	app.AppendReqAndResp(url, "head", handler)
	return app
}

// ANY registers a route that matches all the HTTP methods.
// GET, POST, PUT, HEAD, OPTIONS, DELETE.
// 執行所有方法的AppendReqAndResp(url, "head", handler)，在App(struct)新增Requests([]Path)路徑、Middlewares至Handlers
func (app *App) ANY(url string, handler ...Handler) *App {
	app.routeANY = true
	app.AppendReqAndResp(url, "post", handler)
	app.AppendReqAndResp(url, "get", handler)
	app.AppendReqAndResp(url, "delete", handler)
	app.AppendReqAndResp(url, "put", handler)
	app.AppendReqAndResp(url, "options", handler)
	app.AppendReqAndResp(url, "head", handler)
	return app
}

// 將參數設置至App.Routers(RouterMap)中，設定methods及patten(url)
func (app *App) Name(name string) {
	if app.routeANY {
		// Routers(RouterMap)類別為map[string]Router，Router(struct)裡有methods、patten
		//新增所有方法
		app.Routers[name] = Router{
			Methods: []string{"POST", "GET", "DELETE", "PUT", "OPTIONS", "HEAD"},
			Patten:  app.Requests[app.routeIndex].URL,
		}
	} else {
		app.Routers[name] = Router{
			Methods: []string{app.Requests[app.routeIndex].Method},
			Patten:  app.Requests[app.routeIndex].URL,
		}
	}
}

// Group add middlewares and prefix for App.
// 將參數prefix、middleware新增至RouterGroup(struct)
func (app *App) Group(prefix string, middleware ...Handler) *RouterGroup {
	return &RouterGroup{
		app:         app,
		Middlewares: append(app.Middlewares, middleware...),
		Prefix:      slash(prefix),
	}
}

// RouterGroup is a group of routes.
type RouterGroup struct {
	app         *App //struct
	Middlewares Handlers //Handlers([]Handler)，Handler類別為 func(ctx *Context)
	Prefix      string
}

// AppendReqAndResp stores the request info and handle into app.
// support the route parameter. The route parameter will be recognized as
// wildcard store into the RegUrl of Path struct. For example:
//
//         /user/:id      => /user/(.*)
//         /user/:id/info => /user/(.*?)/info
//
// The RegUrl will be used to recognize the incoming path and find
// the handler.
// 在RouterGroup.app(struct)中新增Requests([]Path)路徑及方法、接著在該url中新增參數handler(Handler...)
func (g *RouterGroup) AppendReqAndResp(url, method string, handler []Handler) {

	g.app.Requests = append(g.app.Requests, Path{
		URL:    join(g.Prefix, url),
		Method: method,
	})
	g.app.routeIndex++

	var h = make([]Handler, len(g.Middlewares))
	copy(h, g.Middlewares)

	g.app.Handlers[Path{
		URL:    join(g.Prefix, url),
		Method: method,
	}] = append(h, handler...)
}

// POST is a shortcut for app.AppendReqAndResp(url, "post", handler).
// POST等於在AppendReqAndResp(url, "post", handler)，在RouterGroup.app(struct)中新增Requests([]Path)路徑及方法、接著在該url中新增參數handler(Handler...)
func (g *RouterGroup) POST(url string, handler ...Handler) *RouterGroup {
	g.app.routeANY = false
	g.AppendReqAndResp(url, "post", handler)
	return g
}

// GET is a shortcut for app.AppendReqAndResp(url, "get", handler).
// GET等於在AppendReqAndResp(url, "get", handler)，在RouterGroup.app(struct)中新增Requests([]Path)路徑及方法、接著在該url中新增參數handler(Handler...)
func (g *RouterGroup) GET(url string, handler ...Handler) *RouterGroup {
	g.app.routeANY = false
	g.AppendReqAndResp(url, "get", handler)
	return g
}

// DELETE is a shortcut for app.AppendReqAndResp(url, "delete", handler).
// DELETE等於在AppendReqAndResp(url, "delete", handler)，在RouterGroup.app(struct)中新增Requests([]Path)路徑及方法、接著在該url中新增參數handler(Handler...)
func (g *RouterGroup) DELETE(url string, handler ...Handler) *RouterGroup {
	g.app.routeANY = false
	g.AppendReqAndResp(url, "delete", handler)
	return g
}

// PUT is a shortcut for app.AppendReqAndResp(url, "put", handler).
// PUT等於在AppendReqAndResp(url, "put", handler)，在RouterGroup.app(struct)中新增Requests([]Path)路徑及方法、接著在該url中新增參數handler(Handler...)
func (g *RouterGroup) PUT(url string, handler ...Handler) *RouterGroup {
	g.app.routeANY = false
	g.AppendReqAndResp(url, "put", handler)
	return g
}

// OPTIONS is a shortcut for app.AppendReqAndResp(url, "options", handler).
// OPTIONS等於在AppendReqAndResp(url, "options", handler)，在RouterGroup.app(struct)中新增Requests([]Path)路徑及方法、接著在該url中新增參數handler(Handler...)
func (g *RouterGroup) OPTIONS(url string, handler ...Handler) *RouterGroup {
	g.app.routeANY = false
	g.AppendReqAndResp(url, "options", handler)
	return g
}

// HEAD is a shortcut for app.AppendReqAndResp(url, "head", handler).
// HEAD等於在AppendReqAndResp(url, "head", handler)，在RouterGroup.app(struct)中新增Requests([]Path)路徑及方法、接著在該url中新增參數handler(Handler...)
func (g *RouterGroup) HEAD(url string, handler ...Handler) *RouterGroup {
	g.app.routeANY = false
	g.AppendReqAndResp(url, "head", handler)
	return g
}

// ANY registers a route that matches all the HTTP methods.
// GET, POST, PUT, HEAD, OPTIONS, DELETE.
// 執行所有方法的AppendReqAndResp(url, "head", handler)，在RouterGroup.app(struct)中新增Requests([]Path)路徑及方法、接著在該url中新增參數handler(Handler...)
func (g *RouterGroup) ANY(url string, handler ...Handler) *RouterGroup {
	g.app.routeANY = true
	g.AppendReqAndResp(url, "post", handler)
	g.AppendReqAndResp(url, "get", handler)
	g.AppendReqAndResp(url, "delete", handler)
	g.AppendReqAndResp(url, "put", handler)
	g.AppendReqAndResp(url, "options", handler)
	g.AppendReqAndResp(url, "head", handler)
	return g
}

// 將參數設置至App.Routers(RouterMap)中，設定methods及patten(url)
func (g *RouterGroup) Name(name string) {
	// RouterGroup.App(struct)的Name方法
	g.app.Name(name)
}

// Group add middlewares and prefix for RouterGroup.
// 將參數prefix、middleware新增至RouterGroup(struct)
func (g *RouterGroup) Group(prefix string, middleware ...Handler) *RouterGroup {
	return &RouterGroup{
		app:         g.app,
		Middlewares: append(g.Middlewares, middleware...),
		Prefix:      join(slash(g.Prefix), slash(prefix)),
	}
}

// slash fix the path which has wrong format problem.
//
// 	 ""      => "/"
// 	 "abc/"  => "/abc"
// 	 "/abc/" => "/abc"
// 	 "/abc"  => "/abc"
// 	 "/"     => "/"
//
// 處理斜線(路徑)
func slash(prefix string) string {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" || prefix == "/" {
		return "/"
	}
	if prefix[0] != '/' {
		if prefix[len(prefix)-1] == '/' {
			return "/" + prefix[:len(prefix)-1]
		}
		return "/" + prefix
	}
	if prefix[len(prefix)-1] == '/' {
		return prefix[:len(prefix)-1]
	}
	return prefix
}

// join join the path.
// join路徑
func join(prefix, suffix string) string {
	if prefix == "/" {
		return suffix
	}
	if suffix == "/" {
		return prefix
	}
	return prefix + suffix
}
