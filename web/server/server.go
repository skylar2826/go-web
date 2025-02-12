package server

import (
	"awesomeProject2/web/context"
	"awesomeProject2/web/filter"
	"awesomeProject2/web/handler"
	template2 "awesomeProject2/web/template"
	"fmt"
	"net/http"
	"sync"
)

type Server interface {
	handler.Routable
	Start(address string) error
	Shutdown() error
}

type SdkHttpServer struct {
	Name    string
	handler handler.Handler
	//root           filter.Filter
	middlewares    []filter.Middleware
	context        *sync.Pool
	templateEngine template2.TemplateEngine
}

func (s *SdkHttpServer) Route(method string, pattern string, handlerFunc handler.HandleFunc) {
	s.handler.Route(method, pattern, handlerFunc)
}

func (s *SdkHttpServer) Start(address string) error {
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		//c := context.NewContext(writer, request)
		c := s.context.Get().(*context.Context)
		defer func() {
			s.context.Put(c)
		}()
		c.Reset(writer, request)
		s.executeMiddlewares(c)
	})
	return http.ListenAndServe(address, nil)
}

func (s *SdkHttpServer) executeMiddlewares(c *context.Context) {
	root := s.handler.ServeHTTP
	for i := len(s.middlewares) - 1; i >= 0; i-- {
		b := s.middlewares[i]
		root = b(root)
	}
	root(c)
}

func (s *SdkHttpServer) Shutdown() error {
	// todo: 关闭服务
	fmt.Printf("%s is closed\n", s.Name)
	return nil
}

//var BuilderMap = make(map[string]filter.Middleware, 4)
//
//func RegisterBuilder(name string, builder filter.Middleware) {
//	BuilderMap[name] = builder
//}
//
//func NewSdkHttpServerWithBuilderName(name string, handler handler.Handler, builderNames ...string) Server {
//	builders := make([]filter.Middleware, 0, len(builderNames))
//	for _, builderName := range builderNames {
//		if builder, ok := BuilderMap[builderName]; !ok {
//			log.Printf("builder 不存在 %s\n", builderName)
//		} else {
//			builders = append(builders, builder)
//		}
//	}
//
//	return NewSdkHttpServer(name, handler, builders...)
//}

type Option func(s *SdkHttpServer)

func WithMiddlewares(middlewares ...filter.Middleware) Option {
	return func(s *SdkHttpServer) {
		s.middlewares = append(s.middlewares, middlewares...)
	}
}

func WithTemplate(engine template2.TemplateEngine) Option {
	return func(s *SdkHttpServer) {
		s.templateEngine = engine
	}
}

func NewSdkHttpServer(name string, handler handler.Handler, Opts ...Option) Server {
	s := &SdkHttpServer{
		Name:        name,
		handler:     handler,
		middlewares: []filter.Middleware{filter.FlashRespBuilder},
	}

	for _, opt := range Opts {
		opt(s)
	}

	if s.templateEngine != nil {
		s.context = &sync.Pool{New: func() interface{} {
			return context.NewEmptyContext(context.WithTemplate(s.templateEngine))
		}}
	} else {
		s.context = &sync.Pool{New: func() interface{} {
			return context.NewEmptyContext()
		}}
	}

	return s
}
