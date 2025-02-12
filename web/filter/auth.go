package filter

import (
	"awesomeProject2/web/context"
	"awesomeProject2/web/session/manager"
	"log"
	"strings"
)

func AuthBuilder(next Filter) Filter {
	return func(c *context.Context) {
		if strings.Contains(c.R.URL.Path, "login") {
			next(c)
			return
		}

		sess, err := manager.WebManager.GetSession(c)
		if err != nil {
			log.Printf("no session err: %s\n", err)
			er := c.UnauthorizedJsonDirect(c.R.URL.Path)
			if er != nil {
				log.Printf("auth err: %s\n", err)
			}
			return
		}
		err = manager.WebManager.RefreshSession(c, sess.ID())
		if err != nil {
			log.Printf("update session err: %s\n", err)
		}

		next(c)
	}
}
