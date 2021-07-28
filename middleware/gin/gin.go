package gin

import (
	"errors"
	"net/http"

	rbns "github.com/n-creativesystem/sdk-go-rbns"
	"github.com/n-creativesystem/sdk-go-rbns/middleware"

	"github.com/gin-gonic/gin"
)

var (
	EreNoSDK = errors.New("client sdk is empty")
)

type GetUserOrganization func(c *gin.Context) (userKey string, organizationName string, err error)

func ginPermissionCheck(c *gin.Context, fn GetUserOrganization, permissionNames ...string) error {
	var client *rbns.Client
	if v, ok := c.Get(rbns.ClientKey); ok {
		client, ok = v.(*rbns.Client)
		if !ok {
			return EreNoSDK
		}
	}
	userKey, organizationName, err := fn(c)
	if err != nil {
		return err
	}
	return middleware.PermissionCheck(client, userKey, organizationName, permissionNames...)
}

func PermissionCheckWithClientOptions(fn GetUserOrganization, permissionNames []string, opts ...rbns.Option) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !clientWithOptions(c, opts...) {
			return
		}
		if err := ginPermissionCheck(c, fn, permissionNames...); err != nil {
			c.AbortWithError(http.StatusForbidden, err)
		} else {
			c.Next()
		}
	}
}

func PermissionCheck(fn GetUserOrganization, permissionNames ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := ginPermissionCheck(c, fn, permissionNames...); err != nil {
			c.AbortWithError(http.StatusForbidden, err)
		} else {
			c.Next()
		}
	}
}
