package fwncs

import (
	"net/http"

	"github.com/n-creativesystem/go-fwncs"
	rbns "github.com/n-creativesystem/go-rbns"
)

func clientWithOptions(c fwncs.Context, opts ...rbns.Option) bool {
	ctx := c.GetContext()
	client, err := rbns.Connection(ctx, opts...)
	if err != nil {
		c.AbortWithStatusAndErrorMessage(http.StatusInternalServerError, err)
		return false
	}
	c.Set(rbns.ClientKey, client)
	return true
}

func ClientWithOptions(opts ...rbns.Option) fwncs.HandlerFunc {
	return func(c fwncs.Context) {
		if !clientWithOptions(c, opts...) {
			return
		}
		c.Next()
	}
}

func Client(client *rbns.Client) fwncs.HandlerFunc {
	return func(c fwncs.Context) {
		c.Set(rbns.ClientKey, client)
		c.Next()
	}
}
