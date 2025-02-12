package filter

import "awesomeProject2/web/context"

type Middleware func(next Filter) Filter

type Filter func(c *context.Context)
