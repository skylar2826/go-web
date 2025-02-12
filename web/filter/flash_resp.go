package filter

import (
	"awesomeProject2/web/context"
	"log"
)

func FlashRespBuilder(next Filter) Filter {
	return func(c *context.Context) {
		next(c)
		c.W.WriteHeader(c.RespStatusCode)
		_, err := c.W.Write(c.RespData)
		if err != nil {
			log.Printf("flashResp err: %s", err.Error())
		}
	}
}
