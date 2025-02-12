package __shutdown

import (
	"awesomeProject2/web/context"
	custom_error "awesomeProject2/web/custom_error"
	"awesomeProject2/web/filter"
	context2 "context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"time"
)

type GracefulShutdown struct {
	reqCnt      int32
	zeroRequest chan struct{}
	closing     int32
}

func (g *GracefulShutdown) GracefulShutdownFilterBuilder(next filter.Filter) filter.Filter {
	return func(c *context.Context) {
		cl := atomic.LoadInt32(&g.closing)
		if cl > 0 {
			// 关闭新的请求
			c.RespStatusCode = http.StatusServiceUnavailable
			//c.W.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		// 旧请求继续处理
		atomic.AddInt32(&g.reqCnt, 1)
		next(c)
		n := atomic.AddInt32(&g.reqCnt, -1)

		// 当前请求为最后一个请求时，关闭请求
		if cl > 0 && n == 0 {
			g.zeroRequest <- struct{}{}
		}
	}
}

// WaitForShutdown 拒绝新请求，处理旧请求，释放资源，关闭服务
// 需要提供超时机制
func WaitForShutdown(hooks ...Hook) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, ShutdownSignals...)
	select {
	case sig := <-signals:
		fmt.Printf("get signal %s, application shutdown\n", sig.String())

		ctx, cancel := context2.WithTimeout(context2.Background(), time.Second*30)
		for _, h := range hooks {
			err := h(ctx)
			if err != nil {
				cancel()
				log.Fatal(custom_error.ErrorTimeout(err.Error()))
			}
		}

		time.AfterFunc(time.Second*5, func() {
			// 退出超时，强制关闭
			cancel()
			log.Fatal(custom_error.ErrorTimeout(""))
		})

		// todo: 正常退出，临时用Exit(0)代替
		cancel()
		os.Exit(0)
	}
}
