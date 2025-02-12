package context

import (
	"awesomeProject2/web/custom_error"
	"awesomeProject2/web/template"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type Context struct {
	W              http.ResponseWriter
	RespStatusCode int
	RespData       []byte
	R              *http.Request
	PathParams     map[string]string
	MatchRoute     string
	templateEngine template.TemplateEngine
	UserValues     map[string]any
}

func (c *Context) ReadJson(val interface{}) error {
	data, err := io.ReadAll(c.R.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, val)
	return err
}

func (c *Context) Reset(w http.ResponseWriter, r *http.Request) {
	c.W = w
	c.R = r
	c.PathParams = make(map[string]string, 1)
}

type CommonResponse struct {
	BizCode int
	Msg     string
	Data    interface{}
}

func (c *Context) WriteResp(res *CommonResponse) error {
	var data []byte
	var err error
	if res.Msg != "" {
		data, err = json.Marshal(res.Msg)
		if err != nil {
			//c.W.WriteHeader(http.StatusServiceUnavailable)
			c.RespStatusCode = http.StatusServiceUnavailable
			return err
		}
	} else {
		data, err = json.Marshal(res.Data)
		if err != nil {
			c.RespStatusCode = http.StatusServiceUnavailable
			//c.W.WriteHeader(http.StatusServiceUnavailable)
			return err
		}
	}
	c.RespStatusCode = res.BizCode
	//c.W.WriteHeader(res.BizCode)
	//_, err = c.W.Write(data)
	c.RespData = data
	return err
}

func (c *Context) BadRequestJson(err error) error {
	res := &CommonResponse{
		BizCode: http.StatusBadRequest,
		Msg:     fmt.Sprintf("request error: %s", err),
	}
	return c.WriteResp(res)
}

func (c *Context) SystemErrorJson(err error) error {
	res := &CommonResponse{
		BizCode: http.StatusInternalServerError,
		Msg:     fmt.Sprintf("system error: %s", err),
	}
	return c.WriteResp(res)
}

func (c *Context) NotFoundJson(str string) error {
	res := &CommonResponse{
		BizCode: http.StatusNotFound,
		Msg:     custom_error.ErrorNotFound(str).Error(),
	}
	return c.WriteResp(res)
}

func (c *Context) UnauthorizedJsonDirect(str string) error {
	c.W.WriteHeader(http.StatusUnauthorized)
	_, err := c.W.Write([]byte(custom_error.ErrorUnauthorizedJson(str).Error()))
	return err
}

func (c *Context) OkJson(data interface{}) error {
	res := &CommonResponse{
		BizCode: http.StatusOK,
		Data:    data,
	}
	return c.WriteResp(res)
}

func (c *Context) OkJsonDirect(data []byte) {
	c.RespStatusCode = http.StatusOK
	c.RespData = data
}

func (c *Context) Render(tplName string, data any) []byte {
	val, err := c.templateEngine.Render(c.R.Context(), tplName, data)
	if err != nil {
		er := c.SystemErrorJson(err)
		if er != nil {
			log.Println("template render err: ", err)
		}
	}
	return val
}

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

func NewEmptyContext(opts ...Option) *Context {
	c := &Context{}
	for _, opt := range opts {
		opt(c)
	}
	return c
}
