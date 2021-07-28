package fwncs

import (
	"n-creativesystem/go-rbns-sdk"
	"n-creativesystem/go-rbns-sdk/middleware"
	"net/http"

	"github.com/n-creativesystem/go-fwncs"
)

type GetUserOrganization func(c fwncs.Context) (userKey string, organizationName string, err error)

func fwncsPermissionCheck(c fwncs.Context, fn GetUserOrganization, permissionNames ...string) error {
	var client *rbns.Client
	if v, ok := c.Get(rbns.ClientKey).(*rbns.Client); ok {
		client = v
	}
	userKey, organizationName, err := fn(c)
	if err != nil {
		return err
	}
	return middleware.PermissionCheck(client, userKey, organizationName, permissionNames...)
}

func PermissionCheckWithClientOptions(fn GetUserOrganization, permissionNames []string, opts ...rbns.Option) fwncs.HandlerFunc {
	return func(c fwncs.Context) {
		if !clientWithOptions(c, opts...) {
			return
		}
		if err := fwncsPermissionCheck(c, fn, permissionNames...); err != nil {
			c.AbortWithStatusAndErrorMessage(http.StatusForbidden, err)
		} else {
			c.Next()
		}
	}
}

func PermissionCheck(fn GetUserOrganization, permissionNames ...string) fwncs.HandlerFunc {
	return func(c fwncs.Context) {
		if err := fwncsPermissionCheck(c, fn, permissionNames...); err != nil {
			c.AbortWithStatusAndErrorMessage(http.StatusForbidden, err)
		} else {
			c.Next()
		}
	}
}
