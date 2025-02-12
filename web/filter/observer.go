package filter

import (
	"awesomeProject2/web/context"
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	trace2 "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"log"
	"os"
	"strconv"
	"time"
)

func init() {
	initZipkin("", "server")
}

func initZipkin(collectorUrl string, serviceName string) {
	exporter, err := zipkin.New(collectorUrl, zipkin.WithLogger(log.New(os.Stderr, "zipkin-server", log.Llongfile)))
	if err != nil {
		log.Println(err.Error())
		return
	}
	b := trace2.NewBatchSpanProcessor(exporter)
	tp := trace2.NewTracerProvider(trace2.WithSpanProcessor(b), trace2.WithResource(resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceNameKey.String(fmt.Sprintf("zipkin-%s", serviceName)))))
	otel.SetTracerProvider(tp)
}

var defaultInstrumentationName = "D:\\workspace\\awesomeProject2\\web\\filter\\observer.go"

type accessLog struct {
	Host       string `json:"Host"`
	MatchRoute string `json:"MatchRoute"`
	Path       string `json:"Path"`
	Method     string `json:"Method"`
}

type ObserverMiddlewareBuilder struct {
	logFunc func(accessLog string)
	tracer  trace.Tracer
	vector  *prometheus.SummaryVec
}

func (m *ObserverMiddlewareBuilder) RegisterLogFunc(logFunc func(accessLog string)) *ObserverMiddlewareBuilder {
	m.logFunc = logFunc
	return m
}

func (m *ObserverMiddlewareBuilder) RegisterVector(Namespace string, Subsystem string, Name string, Help string) *ObserverMiddlewareBuilder {
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: Namespace,
		Subsystem: Subsystem,
		Name:      Name,
		Help:      Help,
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.90:  0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	}, []string{"pattern", "method", "status"})
	m.vector = vector
	prometheus.MustRegister(vector)
	return m
}

func (m *ObserverMiddlewareBuilder) registerTracer() *ObserverMiddlewareBuilder {
	m.tracer = otel.GetTracerProvider().Tracer(defaultInstrumentationName)
	return m
}

func (m *ObserverMiddlewareBuilder) setRequestTracerSpan(c *context.Context) trace.Span {
	if m.tracer == nil {
		return nil
	}

	ctx := c.R.Context()
	ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(c.R.Header))

	var span trace.Span
	ctx, span = m.tracer.Start(ctx, "unknown")

	span.SetAttributes(attribute.String("http.method", c.R.Method))
	span.SetAttributes(attribute.String("peer.hostname", c.R.Host))
	span.SetAttributes(attribute.String("http.url", c.R.URL.String()))
	span.SetAttributes(attribute.String("http.scheme", c.R.URL.Scheme))
	span.SetAttributes(attribute.String("span.kind", "server"))
	span.SetAttributes(attribute.String("component", "web"))
	span.SetAttributes(attribute.String("peer.address", c.R.RemoteAddr))
	span.SetAttributes(attribute.String("http.proto", c.R.Proto))

	c.R = c.R.WithContext(ctx)

	return span
}

func (m *ObserverMiddlewareBuilder) getPattern(c *context.Context) string {
	pattern := c.MatchRoute
	if c.MatchRoute == "" {
		pattern = "unknown"
	}
	return pattern
}

func (m *ObserverMiddlewareBuilder) setResponseTracerSpan(c *context.Context, span trace.Span) {
	if m.tracer == nil {
		return
	}
	pattern := m.getPattern(c)
	span.SetName(pattern)
	span.SetAttributes(attribute.String("http.responseStatus", strconv.Itoa(c.RespStatusCode)))
	span.SetAttributes(attribute.String("http.respData", string(c.RespData)))
}

func (m *ObserverMiddlewareBuilder) observeVector(c *context.Context, startTime time.Time) {
	go func() {
		if m.vector != nil {
			pattern := m.getPattern(c)
			m.vector.WithLabelValues(pattern, c.R.Method, strconv.Itoa(c.RespStatusCode)).Observe(float64(time.Now().Sub(startTime).Milliseconds()))
		}
	}()
}

func (m *ObserverMiddlewareBuilder) log(c *context.Context) {
	l := &accessLog{
		Host:       c.R.Host,
		Path:       c.R.URL.Path,
		Method:     c.R.Method,
		MatchRoute: c.MatchRoute,
	}
	val, err := json.Marshal(l)
	if err != nil {
		m.logFunc(err.Error())
	} else {
		m.logFunc(string(val))
	}
}

func (m *ObserverMiddlewareBuilder) Build() Middleware {
	return func(next Filter) Filter {
		return func(c *context.Context) {
			span := m.setRequestTracerSpan(c)
			startTime := time.Now()
			defer func() {
				if span != nil {
					span.End()
				}

				m.observeVector(c, startTime)
				m.log(c)
			}()
			next(c)
			m.setResponseTracerSpan(c, span)
		}
	}
}

type ObserverOption func(m *ObserverMiddlewareBuilder)

func NewObserverMiddlewareBuilder() *ObserverMiddlewareBuilder {
	m := &ObserverMiddlewareBuilder{
		logFunc: func(accessLog string) {
			fmt.Println(accessLog)
		},
	}

	return m
}
