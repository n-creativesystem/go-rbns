package gin

import (
	"net/http"

	"github.com/n-creativesystem/go-rbns-sdk"

	"github.com/gin-gonic/gin"
)

func clientWithOptions(c *gin.Context, opts ...rbns.Option) bool {
	ctx := c.Request.Context()
	client, err := rbns.Connection(ctx, opts...)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return false
	}
	c.Set(rbns.ClientKey, client)
	return true
}

func ClientWithOptions(opts ...rbns.Option) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !clientWithOptions(c, opts...) {
			return
		}
		c.Next()
	}
}

func Client(client *rbns.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(rbns.ClientKey, client)
		c.Next()
	}
}
