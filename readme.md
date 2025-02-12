# WEB 框架开发



## 基本思想

### 原则

框架提供核心能力，扩展能力以插件方式注册

### 设计模式

#### Option模式

##### 示例

```
type Option func(c *Context)

func WithTemplate(engine template.TemplateEngine) Option {
	return func(c *Context) {
		c.templateEngine = engine
	}
}

func NewContext(w http.ResponseWriter, r *http.Request, opts ...Option) *Context {
	c := &Context{
		W: w,
		R: r,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}
```

#### build模式

##### 示例

```
type ObserverMiddlewareBuilder struct

func (m *ObserverMiddlewareBuilder) Build() Middleware
```

##### 好处

维护type a struct比包方法更灵活

#### 责任链模式

##### 示例

```
type Middleware func(next Filter) Filter

type Filter func(c *context.Context)
```

通过next，像剥洋葱一样调用

## 关键实现

Server定义、路由树、context上下文、AOP方案（Middleware）、静态资源服务与文件处理、页面渲染template、Session、优雅退出

### Server定义

```
type Server interface {
	handler.Routable
	Start(address string) error
	Shutdown() error
}
```

### 路由树

支持静态路由匹配、通配符、参数路径、正则匹配

#### node

```
type Node struct {
	path       string // 用户查找路径节点
	handleFunc handler.HandleFunc
	children   []*Node
	nodeType   int
	pattern    string // route注册匹配规则
	Match      func(path string, c *context.Context) bool
}
```

### context上下文

```
type Context struct {
	W              http.ResponseWriter
	R              *http.Request
	RespStatusCode int
	RespData       []byte
	PathParams     map[string]string // 参数路径中的参数
	MatchRoute     string
	templateEngine template.TemplateEngine
	UserValues     map[string]any // session缓存
}
```

- 将RespStatusCode、RespData维护起来的原因：
  直接ctx.W.Write()后的内容是无法再次修改的；使用RespData维护，可以在server-middleware中修改返回值

- 使用sync.Pool复用context

### AOP方案

使用责任链模式

```
type Middleware func(next Filter) Filter

type Filter func(c *context.Context)
```

#### 可观测性middleware

使用opentelemetry，支持tracer+zipkin、prometheus

```
type ObserverMiddlewareBuilder struct {
	logFunc func(accessLog string)
	tracer  trace.Tracer
	vector  *prometheus.SummaryVec
}
```

### Handler

#### 静态资源服务

使用lru缓存，支持用户扩展支持的extMap

```
type StaticResourceHandler struct {
	dir         string
	pathPrefix  string
	extMap      map[string]string
	c           *lru.Cache
	maxFileSize int
}
```

#### 文件处理

文件上传、下载

### 页面渲染template

使用场景：404时，返回404页面；
使用方式：在server中注入templateEngine; 在route handler中通过ctx.templateEngine中获取自定义使用

```
type TemplateEngine interface {
	Render(ctx context.Context, tplName string, data interface{}) ([]byte, error)
}
```

### Session

```
type Session interface {
	Get(c *context.Context, key string) (interface{}, error)
	Set(c *context.Context, key string, val interface{}) error
	ID() string
}
```

#### Store

```
// Store 管理session
type Store interface {
	Generate(c *context.Context, id string) (Session, error)
	Get(c *context.Context, id string) (Session, error)
	Remove(c *context.Context, id string) error
	Refresh(c *context.Context, id string) error
}
```

##### memoryStore

```
type memorySession struct {
	id  string
	val sync.Map
}
type MemoryStore struct {
	sessions *cache.Cache
	expired  time.Duration
	mutex    sync.Mutex
}
```

##### redisStore

```
type redisSession struct {
	cmd      redis.Cmdable
	id       string
	redisKey string
}

type redisStore struct {
	prefix  string
	cmd     redis.Cmdable
	expired time.Duration
}
```

#### Propagator

Propagator 将session关联http.cookie中

```
type Propagator interface {
	Inject(id string, w http.ResponseWriter) error
	Extract(r *http.Request) (string, error)
	Delete(w http.ResponseWriter) error
}
```

### Manager

Manager 胶水作用，方便用户操作

```
type Manager struct {
	session.Store
	session.Propagator
}
```

### 实现业务

demo实现登录、鉴权

### 优化

减轻redis、数据访问压力：

1. 使用ctx.UserValues缓存使用过的session, 减少访问redis的次数
2. sessionValue中存储非敏感、高频访问数据

### 优雅退出

服务器关闭时：

1. 拒绝新的请求
2. 等待当前的所有请求处理完毕
3. 释放资源
4. 关闭服务器
5. 如果这中间超时了， 我们要强制关闭

```
func WaitForShutdown(hooks ...Hook) {}
使用channel监听关闭信号，监听到关闭信号后，依次执行hook

type Hook func(c context.Context) error

func BuildServerHook(servers ...server.Server) Hook 
```

