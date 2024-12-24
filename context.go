package webber

import (
	"github.com/gin-gonic/gin"
	"github.com/xbmlz/webber/container"
)

type Context struct {
	*container.Container
	*gin.Context
}
