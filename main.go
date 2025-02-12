package main

import (
	"awesomeProject2/web/context"
	file_server "awesomeProject2/web/file-server"
	"awesomeProject2/web/filter"
	__shutdown "awesomeProject2/web/graceful_shutdown"
	"awesomeProject2/web/handler"
	handlebasedontree "awesomeProject2/web/handler/handle_based_on_tree"
	"awesomeProject2/web/handler_func"
	"awesomeProject2/web/server"
	"awesomeProject2/web/session/manager"
	"awesomeProject2/web/template"
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	template2 "html/template"
	"log"
	"net/http"
	"strconv"
	"time"
)

//func init() {
//	server.RegisterBuilder("TimeFilterBuilder", filter.TimeFilterBuilder)
//	server.RegisterBuilder("GracefulShutdownFilterBuilder", __shutdown.G.GracefulShutdownFilterBuilder)
//}

func routeHandler(c *context.Context) {
	sess, err := manager.WebManager.GetSession(c)
	if err != nil {
		_ = c.SystemErrorJson(err)
		return
	}
	var val interface{}
	val, err = sess.Get(c, sess.ID())
	if err != nil {
		_ = c.SystemErrorJson(err)
		return
	}
	user := &handler_func.UserNoSecret{}

	err = json.Unmarshal([]byte(val.(string)), user)
	if err != nil {
		_ = c.SystemErrorJson(err)
		return
	}

	_ = c.OkJson(fmt.Sprintf("hello, %s", user.Name))
}

func routeHandler2(c *context.Context) {
	time.Sleep(time.Second * 4)
	_ = c.OkJson("hello world")
}

func routeHandler3(c *context.Context) {
	// 假设这里找不到,返回404
	val := c.Render(strconv.Itoa(http.StatusNotFound), "mock 404")
	c.RespData = val
	c.RespStatusCode = http.StatusNotFound
}

func routeHandler4(c *context.Context) {
	val := c.Render(strconv.Itoa(http.StatusInternalServerError), "mock 500")
	c.RespData = val
	c.RespStatusCode = http.StatusInternalServerError
}

func getStaticResourceHandler() *handler.StaticResourceHandler {
	staticResourceCache := handler.WithFileCache(100000, 2)
	staticResourceExt := handler.WithMoreStaticResourceExt(map[string]string{"html": "text/html"})
	return handler.NewStaticResourceHandler("D:/workspace/awesomeProject2/web/static", "/front", staticResourceExt, staticResourceCache)
}

func NewWebServer() {
	// 注册404、500页面模板
	tpl, err := template2.ParseGlob("web/template/*.gohtml")
	if err != nil {
		log.Println("template加载出错", err)
	}
	goTemplate := &template.GoTemplateEngine{Template: tpl}

	// 使用路由树结构
	hdl := handlebasedontree.NewHandleBasedOnTree()
	//s := server.NewSdkHttpServerWithBuilderName("web-s", hdl, "GracefulShutdownFilterBuilder", "test2")

	// 数据监控
	middlewareBuilder := filter.NewObserverMiddlewareBuilder().RegisterVector("app", "web_server", "vector", "")
	// java -jar zipkin.jar 启动zipkin
	// zipkin 位置：D:\workspace\zipkin
	observer := middlewareBuilder.Build()

	s := server.NewSdkHttpServer("web-server", hdl, server.WithMiddlewares(observer, filter.AuthBuilder), server.WithTemplate(goTemplate))
	//s := server.NewSdkHttpServer("web-server", hdl, filter.TimeFilterBuilder, __shutdown.G.GracefulShutdownFilterBuilder, observer)
	s.Route(http.MethodGet, "/", routeHandler)
	//// 不存在节点
	s.Route(http.MethodGet, "/a", routeHandler2) // sleep 4s
	// 存在一部分且没孩子
	s.Route(http.MethodGet, "/a/b", routeHandler3) // 验证模板404
	//// 存在一部分且有孩子
	s.Route(http.MethodGet, "/a/d", routeHandler4) // 验证模板500
	//// 不存在长路径
	//s.Route(http.MethodGet, "/c/f/g/h", routeHandler4)
	//s.Route(http.MethodPost, "/signUp", handlerFunc.SignUp)
	//s.Route(http.MethodGet, "/a/*", routeHandler2)
	//s.Route(http.MethodGet, "/a/*/b", routeHandler2)

	//s.Route(http.MethodGet, "/a/b/c", routeHandler2)
	//s.Route(http.MethodGet, "/b/*", routeHandler2)
	//s.Route(http.MethodGet, "/c/:id", routeHandler3)
	//s.Route(http.MethodGet, "/d/[1-9]", routeHandler4)

	s.Route(http.MethodPost, "/login", handler_func.SignUp)

	// 静态资源服务
	staticResourceHandler := getStaticResourceHandler()
	s.Route(http.MethodGet, "/front/*", staticResourceHandler.ServeStaticResource)

	// 上传下载服务

	uploader := file_server.NewFileUploader("web/testdata/file", "myFile")
	s.Route(http.MethodPost, "/upload", uploader.Handle())

	downloader := file_server.NewFileDownloader("web/testdata/file", "file")
	s.Route(http.MethodGet, "/download", downloader.Handle())

	// 数据监控
	go func() {
		http.Handle("/metics", promhttp.Handler())
		_ = http.ListenAndServe("127.0.0.1:8082", nil)
	}()

	// 拒绝请求，关闭服务，释放资源
	go func() {
		__shutdown.WaitForShutdown(__shutdown.RejectRequestHook, __shutdown.BuildServerHook(s))
	}()

	_ = s.Start("127.0.0.1:8081")
}

func main() {
	NewWebServer()
}
