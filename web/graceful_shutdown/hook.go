package __shutdown

import (
	customerror "awesomeProject2/web/custom_error"
	"awesomeProject2/web/server"
	"context"
	"fmt"
	"sync"
	"sync/atomic"
)

type Hook func(c context.Context) error

var G = &GracefulShutdown{}

// RejectRequestHook 发信号，通知关闭新请求， 处理完成已有请求
func RejectRequestHook(c context.Context) error {
	atomic.AddInt32(&G.closing, 1)

	reqCnt := atomic.LoadInt32(&G.reqCnt)
	if reqCnt == 0 {
		return nil
	}

	select {
	case <-c.Done():
		return customerror.ErrorShutDownServer("request handle timeout")
	case <-G.zeroRequest:
		fmt.Printf("request hanlde fininshed")
	}
	return nil
}

// BuildServerHook 关闭服务
func BuildServerHook(servers ...server.Server) Hook {
	return func(c context.Context) error {
		wg := &sync.WaitGroup{}
		doneCh := make(chan struct{})
		wg.Add(len(servers))
		for _, s := range servers {
			go func(s server.Server) {
				err := s.Shutdown()
				if err != nil {
					fmt.Printf(err.Error())
				}
				wg.Done()
			}(s)
		}

		go func() {
			wg.Wait()
			doneCh <- struct{}{}
		}()

		select {
		case <-c.Done():
			return customerror.ErrorTimeout("servers handle timeout")
		case <-doneCh:
			fmt.Printf("servers closed fininshed.\n")
		}

		return nil
	}
}
